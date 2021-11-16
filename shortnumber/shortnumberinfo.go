package shortnumber

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
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

type ShortNumberCost int

const (
	UNKNOWN_COST ShortNumberCost = 0
	PREMIUM_RATE                 = iota
	TOLL_FREE
	STANDARD_RATE
)

func GetExpectedCostForRegion(number *phonenumbers.PhoneNumber, regionDialingFrom string) ShortNumberCost {
	if !regionDialingFromMatchesNumber(number, regionDialingFrom) {
		return UNKNOWN_COST
	}

	phoneMetadata := getMetadataForRegion(regionDialingFrom)
	if phoneMetadata == nil {
		return UNKNOWN_COST
	}

	shortNumber := getNationalSignificantNumber(number)

	if !uint32ListContains(phoneMetadata.GetGeneralDesc().GetPossibleLength(), int32(len(shortNumber))) {
		return UNKNOWN_COST
	}

	if matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.GetPremiumRate()) {
		return PREMIUM_RATE
	}
	if matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.GetStandardRate()) {
		return STANDARD_RATE
	}
	if matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.GetTollFree()) {
		fmt.Printf("toll free: %+v\n", phoneMetadata.GetTollFree())
		return TOLL_FREE
	}
	if IsEmergencyNumber(shortNumber, regionDialingFrom) {
		fmt.Println("emergency?")
		// Emergency numbers are implicitly toll-free.
		return TOLL_FREE
	}

	return UNKNOWN_COST
}

func uint32ListContains(l []int32, e int32) bool {
	for _, i := range l {
		if i == e {
			return true
		}
	}
	return false
}

/*
  public ShortNumberCost getExpectedCostForRegion(PhoneNumber number, String regionDialingFrom) {
    if (!regionDialingFromMatchesNumber(number, regionDialingFrom)) {
      return ShortNumberCost.UNKNOWN_COST;
    }
    // Note that regionDialingFrom may be null, in which case phoneMetadata will also be null.
    PhoneMetadata phoneMetadata = MetadataManager.getShortNumberMetadataForRegion(
        regionDialingFrom);
    if (phoneMetadata == null) {
      return ShortNumberCost.UNKNOWN_COST;
    }

    String shortNumber = getNationalSignificantNumber(number);

    // The possible lengths are not present for a particular sub-type if they match the general
    // description; for this reason, we check the possible lengths against the general description
    // first to allow an early exit if possible.
    if (!phoneMetadata.getGeneralDesc().getPossibleLengthList().contains(shortNumber.length())) {
      return ShortNumberCost.UNKNOWN_COST;
    }

    // The cost categories are tested in order of decreasing expense, since if for some reason the
    // patterns overlap the most expensive matching cost category should be returned.
    if (matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.getPremiumRate())) {
      return ShortNumberCost.PREMIUM_RATE;
    }
    if (matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.getStandardRate())) {
      return ShortNumberCost.STANDARD_RATE;
    }
    if (matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.getTollFree())) {
      return ShortNumberCost.TOLL_FREE;
    }
    if (isEmergencyNumber(shortNumber, regionDialingFrom)) {
      // Emergency numbers are implicitly toll-free.
      return ShortNumberCost.TOLL_FREE;
    }
    return ShortNumberCost.UNKNOWN_COST;
  }
*/

func matchesPossibleNumberAndNationalNumber(number string, numberDesc *phonenumbers.PhoneNumberDesc) bool {
	if len(numberDesc.GetPossibleLength()) > 0 && !uint32ListContains(numberDesc.GetPossibleLength(), int32(len(number))) {
		return false
	}
	foo := matchNationalNumber(number, numberDesc, false)
	fmt.Printf("matchesPossibleNumberAndNationalNumber(%s...) = %+v\n", number, foo)
	return foo
}

/*
private boolean matchesPossibleNumberAndNationalNumber(String number,
PhoneNumberDesc numberDesc) {
	if (numberDesc.getPossibleLengthCount() > 0
	&& !numberDesc.getPossibleLengthList().contains(number.length())) {
		return false;
	}
	return matcherApi.matchNationalNumber(number, numberDesc, false);
}
*/

func matchNationalNumber(number string, numberDesc *phonenumbers.PhoneNumberDesc, allowPrefixMatch bool) bool {
	fmt.Printf("matchNationalNumber(%s..)\n", number)
	nationalNumberPattern := numberDesc.GetNationalNumberPattern()
	// We don't want to consider it a prefix match when matching non-empty input against an empty
	// pattern.
	if len(nationalNumberPattern) == 0 {
		return false
	}

	if !allowPrefixMatch {
		nationalNumberPattern = "^(?:" + nationalNumberPattern + ")$" // Strictly match
	}
	fmt.Printf("nationalNumberPattern = %+v\n", nationalNumberPattern)
	pattern := regexFor(nationalNumberPattern)

	foo := pattern.MatchString(number)
	fmt.Printf("MatchString(%+v) = %+v\n", number, foo)
	return foo
}

/*
public boolean matchNationalNumber(CharSequence number, PhoneNumberDesc numberDesc,
boolean allowPrefixMatch) {
	String nationalNumberPattern = numberDesc.getNationalNumberPattern();
	// We don't want to consider it a prefix match when matching non-empty input against an empty
	// pattern.
	if (nationalNumberPattern.length() == 0) {
		return false;
	}
	return match(number, regexCache.getPatternForRegex(nationalNumberPattern), allowPrefixMatch);
}

private static boolean match(CharSequence number, Pattern pattern, boolean allowPrefixMatch) {
	Matcher matcher = pattern.matcher(number);
	if (!matcher.lookingAt()) {
		return false;
	} else {
		return (matcher.matches()) ? true : allowPrefixMatch;
	}
}
*/

func regionDialingFromMatchesNumber(number *phonenumbers.PhoneNumber, regionDialingFrom string) bool {
	regionCodes := getRegionCodesForCountryCode(int(number.GetCountryCode()))
	for _, regionCode := range regionCodes {
		if regionCode == regionDialingFrom {
			return true
		}
	}
	return false
}

/*
private boolean regionDialingFromMatchesNumber(PhoneNumber number,
String regionDialingFrom) {
	List<String> regionCodes = getRegionCodesForCountryCode(number.getCountryCode());
	return regionCodes.contains(regionDialingFrom);
}
*/

func getRegionCodesForCountryCode(countryCallingCode int) []string {
	return countryCodeToRegion[countryCallingCode]
}

/*
  private List<String> getRegionCodesForCountryCode(int countryCallingCode) {
    List<String> regionCodes = countryCallingCodeToRegionCodeMap.get(countryCallingCode);
    return Collections.unmodifiableList(regionCodes == null ? new ArrayList<String>(0)
                                                            : regionCodes);
  }
*/

func getNationalSignificantNumber(number *phonenumbers.PhoneNumber) string {
	var nationalNumber strings.Builder
	if number.GetItalianLeadingZero() && number.GetNumberOfLeadingZeros() > 0 {
		nationalNumber.WriteString(strings.Repeat("0", int(number.GetNumberOfLeadingZeros())))
	}
	nationalNumber.WriteString(strconv.FormatUint(number.GetNationalNumber(), 10))
	return nationalNumber.String()
}

/*
public String getNationalSignificantNumber(PhoneNumber number) {
	// If leading zero(s) have been set, we prefix this now. Note this is not a national prefix.
	StringBuilder nationalNumber = new StringBuilder();
	if (number.isItalianLeadingZero() && number.getNumberOfLeadingZeros() > 0) {
		char[] zeros = new char[number.getNumberOfLeadingZeros()];
		Arrays.fill(zeros, '0');
		nationalNumber.append(new String(zeros));
	}
	nationalNumber.append(number.getNationalNumber());
	return nationalNumber.toString();
}
*/

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

	fmt.Printf("metadata: %+v\n", metadata.GetEmergency())
	natRulePattern := metadata.GetEmergency().GetNationalNumberPattern()
	if !allowPrefixMatchForRegion {
		natRulePattern = "^(?:" + natRulePattern + ")$" // Strictly match
	}
	fmt.Printf("natRulePattern = %+v\n", natRulePattern)
	nationalNumberRule := regexFor(natRulePattern)

	foo := nationalNumberRule.MatchString(normalizedNumber)
	fmt.Printf("matchesEmergencyNumberHelper(%s, %s, %+v) = %+v\n", number, regionCode, allowPrefixMatch, foo)
	return foo
}

/*
  private boolean matchesEmergencyNumberHelper(CharSequence number, String regionCode,
  boolean allowPrefixMatch) {
	  CharSequence possibleNumber = PhoneNumberUtil.extractPossibleNumber(number);
	  if (PhoneNumberUtil.PLUS_CHARS_PATTERN.matcher(possibleNumber).lookingAt()) {
		  // Returns false if the number starts with a plus sign. We don't believe dialing the country
		  // code before emergency numbers (e.g. +1911) works, but later, if that proves to work, we can
		  // add additional logic here to handle it.
		  return false;
	  }
	  PhoneMetadata metadata = MetadataManager.getShortNumberMetadataForRegion(regionCode);
	  if (metadata == null || !metadata.hasEmergency()) {
		  return false;
	  }

	  String normalizedNumber = PhoneNumberUtil.normalizeDigitsOnly(possibleNumber);
	  boolean allowPrefixMatchForRegion =
	  allowPrefixMatch && !REGIONS_WHERE_EMERGENCY_NUMBERS_MUST_BE_EXACT.contains(regionCode);
	  return matcherApi.matchNationalNumber(normalizedNumber, metadata.getEmergency(),
	  allowPrefixMatchForRegion);
  }

*/

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
