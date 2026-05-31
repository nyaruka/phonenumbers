// Port of metadata/source/* + MetadataLoader from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
// Last reconciled against: v9.0.32
package phonenumbers

import "google.golang.org/protobuf/proto"

func initMetadata() {
	// load our regions
	regionMap, err := loadIntArrayMap(regionData)
	if err != nil {
		panic(err)
	}
	countryCodeToRegion = regionMap.Map

	// then our metadata
	if err = loadMetadataFromFile(); err != nil {
		panic(err)
	}

	for eKey, regionCodes := range countryCodeToRegion {
		// We can assume that if the county calling code maps to the
		// non-geo entity region code then that's the only region code
		// it maps to.
		if len(regionCodes) == 1 && REGION_CODE_FOR_NON_GEO_ENTITY == regionCodes[0] {
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
	delete(supportedRegions, REGION_CODE_FOR_NON_GEO_ENTITY)

	for _, val := range countryCodeToRegion[NANPA_COUNTRY_CODE] {
		writeToNanpaRegions(val, struct{}{})
	}
}

var (
	// The set of regions that share country calling code 1.
	// There are roughly 26 regions.
	nanpaRegions = make(map[string]struct{})

	// A mapping from a region code to the PhoneMetadata for that region.
	// Note: Synchronization, though only needed for the Android version
	// of the library, is used in all versions for consistency.
	regionToMetadataMap = make(map[string]*PhoneMetadata)

	// A mapping from a country calling code for a non-geographical
	// entity to the PhoneMetadata for that country calling code.
	// Examples of the country calling codes include 800 (International
	// Toll Free Service) and 808 (International Shared Cost Service).
	// Note: Synchronization, though only needed for the Android version
	// of the library, is used in all versions for consistency.
	countryCodeToNonGeographicalMetadataMap = make(map[int]*PhoneMetadata)

	// The set of regions the library supports.
	// There are roughly 240 of them and we set the initial capacity of
	// the HashSet to 320 to offer a load factor of roughly 0.75.
	supportedRegions = make(map[string]bool, 320)

	// The set of calling codes that map to the non-geo entity
	// region ("001"). This set currently contains < 12 elements so the
	// default capacity of 16 (load factor=0.75) is fine.
	countryCodesForNonGeographicalRegion = make(map[int]bool, 16)

	// All the calling codes we support
	supportedCallingCodes = make(map[int]bool, 320)

	// Our map from country code (as integer) to two letter region codes
	countryCodeToRegion map[int][]string
)

func readFromNanpaRegions(key string) (struct{}, bool) {
	v, ok := nanpaRegions[key]
	return v, ok
}

func writeToNanpaRegions(key string, val struct{}) {
	nanpaRegions[key] = val
}

func readFromRegionToMetadataMap(key string) (*PhoneMetadata, bool) {
	v, ok := regionToMetadataMap[key]
	return v, ok
}

func writeToRegionToMetadataMap(key string, val *PhoneMetadata) {
	regionToMetadataMap[key] = val
}

func readFromCountryCodeToNonGeographicalMetadataMap(key int) (*PhoneMetadata, bool) {
	v, ok := countryCodeToNonGeographicalMetadataMap[key]
	return v, ok
}

func writeToCountryCodeToNonGeographicalMetadataMap(key int, v *PhoneMetadata) {
	countryCodeToNonGeographicalMetadataMap[key] = v
}

func loadMetadataFromFile() error {
	metadataCollection, err := MetadataCollection()
	if err != nil {
		return err
	} else if currMetadataColl == nil {
		currMetadataColl = metadataCollection
	}

	metadataList := metadataCollection.GetMetadata()
	if len(metadataList) == 0 {
		return ErrEmptyMetadata
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
	currMetadataColl *PhoneMetadataCollection
	reloadMetadata   = true
)

func MetadataCollection() (*PhoneMetadataCollection, error) {
	if !reloadMetadata {
		return currMetadataColl, nil
	}

	rawBytes, err := decodeUnzip(numberData)
	if err != nil {
		return nil, err
	}

	var metadataCollection = &PhoneMetadataCollection{}
	err = proto.Unmarshal(rawBytes, metadataCollection)
	reloadMetadata = false
	return metadataCollection, err
}
