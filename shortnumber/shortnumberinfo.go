package shortnumber

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/nyaruka/phonenumbers"
)

var (
	// golang map is not go routine safe. Sometimes process exiting
	// because of panic. So adding mutex to synchronize the operation.

	// The set of regions that share country calling code 1.
	// There are roughly 26 regions.
	nanpaRegions = make(map[string]struct{})

	// A mapping from a region code to the PhoneMetadata for that region.
	// Note: Synchronization, though only needed for the Android version
	// of the library, is used in all versions for consistency.
	regionToMetadataMap = make(map[string]*phonenumbers.PhoneMetadata)

	// A mapping from a country calling code for a non-geographical
	// entity to the PhoneMetadata for that country calling code.
	// Examples of the country calling codes include 800 (International
	// Toll Free Service) and 808 (International Shared Cost Service).
	// Note: Synchronization, though only needed for the Android version
	// of the library, is used in all versions for consistency.
	countryCodeToNonGeographicalMetadataMap = make(map[int]*phonenumbers.PhoneMetadata)

	// A cache for frequently used region-specific regular expressions.
	// The initial capacity is set to 100 as this seems to be an optimal
	// value for Android, based on performance measurements.
	regexCache    = make(map[string]*regexp.Regexp)
	regCacheMutex sync.RWMutex

	// The set of regions the library supports.
	// There are roughly 240 of them and we set the initial capacity of
	// the HashSet to 320 to offer a load factor of roughly 0.75.
	supportedRegions = make(map[string]bool, 320)

	// The set of calling codes that map to the non-geo entity
	// region ("001"). This set currently contains < 12 elements so the
	// default capacity of 16 (load factor=0.75) is fine.
	countryCodesForNonGeographicalRegion = make(map[int]bool, 16)

	// These are our onces and maps for our prefix to carrier maps
	carrierOnces     = make(map[string]*sync.Once)
	carrierPrefixMap = make(map[string]*intStringMap)

	// These are our onces and maps for our prefix to geocoding maps
	geocodingOnces     = make(map[string]*sync.Once)
	geocodingPrefixMap = make(map[string]*intStringMap)

	// All the calling codes we support
	supportedCallingCodes = make(map[int]bool, 320)

	// Our once and map for prefix to timezone lookups
	timezoneOnce sync.Once
	// timezoneMap  *intStringArrayMap

	// Our map from country code (as integer) to two letter region codes
	countryCodeToRegion map[int][]string
)

var (
	REGIONS_WHERE_EMERGENCY_NUMBERS_MUST_BE_EXACT map[string]bool = map[string]bool{
		"BR": true,
		"CL": true,
		"NI": true,
	}
)

func init() {
	// load our regions
	regionMap, err := loadIntStringArrayMap(regionMapData)
	if err != nil {
		panic(err)
	}
	countryCodeToRegion = regionMap.Map

	// then our metadata
	err = loadMetadataFromFile("US", 1)
	if err != nil {
		panic(err)
	}

	for eKey, regionCodes := range countryCodeToRegion {
		// We can assume that if the county calling code maps to the
		// non-geo entity region code then that's the only region code
		// it maps to.
		if len(regionCodes) == 1 && phonenumbers.REGION_CODE_FOR_NON_GEO_ENTITY == regionCodes[0] {
			// This is the subset of all country codes that map to the
			// non-geo entity region code.
			countryCodesForNonGeographicalRegion[eKey] = true
		} else {
			// The supported regions set does not include the "001"
			// non-geo entity region code.
			for _, val := range regionCodes {
				supportedRegions[val] = true
			}
		}

		supportedCallingCodes[eKey] = true
	}
	// If the non-geo entity still got added to the set of supported
	// regions it must be because there are entries that list the non-geo
	// entity alongside normal regions (which is wrong). If we discover
	// this, remove the non-geo entity from the set of supported regions
	// and log (or not log).
	delete(supportedRegions, phonenumbers.REGION_CODE_FOR_NON_GEO_ENTITY)

	for _, val := range countryCodeToRegion[phonenumbers.NANPA_COUNTRY_CODE] {
		writeToNanpaRegions(val, struct{}{})
	}

	// Create our sync.Onces for each of our languages for carriers
	// TODO Do we need this?
	// for lang := range carrierMapData {
	// 	carrierOnces[lang] = &sync.Once{}
	// }
	// for lang := range geocodingMapData {
	// 	geocodingOnces[lang] = &sync.Once{}
	// }
}

func writeToNanpaRegions(key string, val struct{}) {
	nanpaRegions[key] = val
}

func loadMetadataFromFile(
	regionCode string,
	countryCallingCode int) error {

	metadataCollection, err := MetadataCollection()
	if err != nil {
		return err
	} else if currMetadataColl == nil {
		currMetadataColl = metadataCollection
	}

	metadataList := metadataCollection.GetMetadata()
	if len(metadataList) == 0 {
		return phonenumbers.ErrEmptyMetadata
	}

	for _, meta := range metadataList {
		region := meta.GetId()
		if region == "001" {
			// it's a non geographical entity
			writeToCountryCodeToNonGeographicalMetadataMap(int(meta.GetCountryCode()), meta)
		} else {
			writeToRegionToMetadataMap(region, meta)
		}
	}
	return nil
}

var (
	currMetadataColl *phonenumbers.PhoneMetadataCollection
	reloadMetadata   = true
)

func MetadataCollection() (*phonenumbers.PhoneMetadataCollection, error) {
	if !reloadMetadata {
		return currMetadataColl, nil
	}

	rawBytes, err := decodeUnzipString(metadataData)
	if err != nil {
		return nil, err
	}

	var metadataCollection = &phonenumbers.PhoneMetadataCollection{}
	err = proto.Unmarshal(rawBytes, metadataCollection)
	reloadMetadata = false
	return metadataCollection, err
}

func loadIntStringArrayMap(data string) (*intStringArrayMap, error) {
	rawBytes, err := decodeUnzipString(data)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(rawBytes)

	// ok, first read in our number of values
	var valueSize uint32
	err = binary.Read(reader, binary.LittleEndian, &valueSize)
	if err != nil {
		return nil, err
	}

	// then our values
	valueBytes := make([]byte, valueSize)
	n, err := reader.Read(valueBytes)
	if uint32(n) < valueSize {
		return nil, fmt.Errorf("unable to read all values: %v", err)
	}

	values := strings.Split(string(valueBytes), "\n")

	// read our # of mappings
	var mappingCount uint32
	err = binary.Read(reader, binary.LittleEndian, &mappingCount)
	if err != nil {
		return nil, err
	}

	maxLength := 0
	mappings := make(map[int][]string, mappingCount)
	key := 0
	for i := 0; i < int(mappingCount); i++ {
		// first read our diff
		diff, err := binary.ReadUvarint(reader)
		if err != nil {
			return nil, err
		}

		key += int(diff)

		// then our values
		var valueCount uint8
		if err = binary.Read(reader, binary.LittleEndian, &valueCount); err != nil {
			return nil, err
		}

		keyValues := make([]string, valueCount)
		for i := 0; i < int(valueCount); i++ {
			var valueIntern uint16
			err = binary.Read(reader, binary.LittleEndian, &valueIntern)
			if err != nil || int(valueIntern) >= len(values) {
				return nil, fmt.Errorf("unable to read interned value: %v", err)
			}
			keyValues[i] = values[valueIntern]
		}
		mappings[key] = keyValues

		strPrefix := fmt.Sprintf("%d", key)
		if len(strPrefix) > maxLength {
			maxLength = len(strPrefix)
		}
	}

	// return our values
	return &intStringArrayMap{
		Map:       mappings,
		MaxLength: maxLength,
	}, nil
}

func ConnectsToEmergencyNumber(number string, regionCode string) bool {
	return matchesEmergencyNumberHelper(number, regionCode, true /* allows prefix match */)
}

func IsEmergencyNumber(number string, regionCode string) bool {
	return matchesEmergencyNumberHelper(number, regionCode, false /* doesn't allow prefix match */)
}

func matchesEmergencyNumberHelper(number string, regionCode string, allowPrefixMatch bool) bool {
	possibleNumber := extractPossibleNumber(number)

	if phonenumbers.PLUS_CHARS_PATTERN.MatchString(possibleNumber) {
		// Returns false if the number starts with a plus sign. We don't believe dialing the country
		// code before emergency numbers (e.g. +1911) works, but later, if that proves to work, we can
		// add additional logic here to handle it.
		return false
	}

	metadata := getMetadataForRegion(regionCode)

	if metadata == nil {
		return false
	}

	normalizedNumber := phonenumbers.NormalizeDigitsOnly(number)

	allowPrefixMatchForRegion := allowPrefixMatch && !REGIONS_WHERE_EMERGENCY_NUMBERS_MUST_BE_EXACT[regionCode]

	natRulePattern := metadata.GetGeneralDesc().GetNationalNumberPattern()
	if !allowPrefixMatchForRegion {
		natRulePattern = "^(?:" + natRulePattern + ")$" // Strictly match
	}
	nationalNumberRule := regexFor(natRulePattern)

	return nationalNumberRule.MatchString(normalizedNumber)
}

// Returns the metadata for the given region code or nil if the region
// code is invalid or unknown.
func getMetadataForRegion(regionCode string) *phonenumbers.PhoneMetadata {
	if !isValidRegionCode(regionCode) {
		return nil
	}
	val, _ := readFromRegionToMetadataMap(regionCode)
	return val
}

// Helper function to check region code is not unknown or null.
func isValidRegionCode(regionCode string) bool {
	valid := supportedRegions[regionCode]
	return len(regionCode) != 0 && valid
}

func readFromCountryCodeToNonGeographicalMetadataMap(key int) (*phonenumbers.PhoneMetadata, bool) {
	v, ok := countryCodeToNonGeographicalMetadataMap[key]
	return v, ok
}

func writeToCountryCodeToNonGeographicalMetadataMap(key int, v *phonenumbers.PhoneMetadata) {
	countryCodeToNonGeographicalMetadataMap[key] = v
}

func readFromRegionToMetadataMap(key string) (*phonenumbers.PhoneMetadata, bool) {
	v, ok := regionToMetadataMap[key]
	return v, ok
}

func writeToRegionToMetadataMap(key string, val *phonenumbers.PhoneMetadata) {
	regionToMetadataMap[key] = val
}
