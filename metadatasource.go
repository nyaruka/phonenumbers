// Port of metadata/source/* + MetadataLoader from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
// Last reconciled against: v9.0.32
package phonenumbers

import "google.golang.org/protobuf/proto"

// metadataContainer bundles all of the metadata-derived lookup state used by
// the library. The package keeps a single active container (currentMetadata),
// built from the embedded metadata during init. Bundling the state behind one
// swappable value is the seam that lets tests run against alternate metadata
// (e.g. upstream's synthetic PhoneNumberMetadataForTesting data) without
// changing the public, package-level API.
type metadataContainer struct {
	// The parsed metadata collection this container was built from.
	metadataCollection *PhoneMetadataCollection

	// A map from country code (as integer) to two letter region codes.
	countryCodeToRegion map[int][]string

	// A mapping from a region code to the PhoneMetadata for that region.
	regionToMetadataMap map[string]*PhoneMetadata

	// A mapping from a country calling code for a non-geographical entity to
	// the PhoneMetadata for that country calling code. Examples of the country
	// calling codes include 800 (International Toll Free Service) and 808
	// (International Shared Cost Service).
	countryCodeToNonGeographicalMetadataMap map[int]*PhoneMetadata

	// The set of regions that share country calling code 1.
	nanpaRegions map[string]struct{}

	// The set of regions the library supports.
	supportedRegions map[string]bool

	// All the calling codes we support.
	supportedCallingCodes map[int]bool

	// The set of calling codes that map to the non-geo entity region ("001").
	countryCodesForNonGeographicalRegion map[int]bool
}

// currentMetadata is the active metadata container. It is populated from the
// embedded metadata by initMetadata during package initialization. Tests may
// swap it via useMetadata to exercise the library against alternate metadata.
var currentMetadata *metadataContainer

func initMetadata() {
	// load our regions
	regionMap, err := loadIntArrayMap(regionData)
	if err != nil {
		panic(err)
	}

	// then our metadata
	coll, err := MetadataCollection()
	if err != nil {
		panic(err)
	}

	currentMetadata, err = newMetadataContainer(coll, regionMap.Map)
	if err != nil {
		panic(err)
	}
}

// newMetadataContainer builds a fully populated metadataContainer from a parsed
// metadata collection and a country-code-to-region map. This is the same
// derivation the library has always performed at init, now factored out so it
// can run against any metadata source, not just the embedded one.
func newMetadataContainer(coll *PhoneMetadataCollection, countryCodeToRegion map[int][]string) (*metadataContainer, error) {
	metadataList := coll.GetMetadata()
	if len(metadataList) == 0 {
		return nil, ErrEmptyMetadata
	}

	mc := &metadataContainer{
		metadataCollection:                      coll,
		countryCodeToRegion:                     countryCodeToRegion,
		regionToMetadataMap:                     make(map[string]*PhoneMetadata),
		countryCodeToNonGeographicalMetadataMap: make(map[int]*PhoneMetadata),
		nanpaRegions:                            make(map[string]struct{}),
		supportedRegions:                        make(map[string]bool, 320),
		supportedCallingCodes:                   make(map[int]bool, 320),
		countryCodesForNonGeographicalRegion:    make(map[int]bool, 16),
	}

	for _, meta := range metadataList {
		region := meta.GetId()
		if region == "001" {
			// it's a non geographical entity
			mc.countryCodeToNonGeographicalMetadataMap[int(meta.GetCountryCode())] = meta
		} else {
			mc.regionToMetadataMap[region] = meta
		}
	}

	for eKey, regionCodes := range countryCodeToRegion {
		// We can assume that if the county calling code maps to the
		// non-geo entity region code then that's the only region code
		// it maps to.
		if len(regionCodes) == 1 && REGION_CODE_FOR_NON_GEO_ENTITY == regionCodes[0] {
			// This is the subset of all country codes that map to the
			// non-geo entity region code.
			mc.countryCodesForNonGeographicalRegion[eKey] = true
		} else {
			// The supported regions set does not include the "001"
			// non-geo entity region code.
			for _, val := range regionCodes {
				mc.supportedRegions[val] = true
			}
		}

		mc.supportedCallingCodes[eKey] = true
	}
	// If the non-geo entity still got added to the set of supported regions it
	// must be because there are entries that list the non-geo entity alongside
	// normal regions (which is wrong). If we discover this, remove the non-geo
	// entity from the set of supported regions.
	delete(mc.supportedRegions, REGION_CODE_FOR_NON_GEO_ENTITY)

	for _, val := range countryCodeToRegion[NANPA_COUNTRY_CODE] {
		mc.nanpaRegions[val] = struct{}{}
	}

	return mc, nil
}

// useMetadata swaps the active metadata container, returning a function that
// restores the previously active container. It is intended for tests that need
// to run against alternate (e.g. synthetic) metadata; callers must invoke the
// returned restore function (typically via t.Cleanup) and must not run such
// tests in parallel, since the active container is process-global.
func useMetadata(mc *metadataContainer) (restore func()) {
	prev := currentMetadata
	currentMetadata = mc
	return func() { currentMetadata = prev }
}

func readFromNanpaRegions(key string) (struct{}, bool) {
	v, ok := currentMetadata.nanpaRegions[key]
	return v, ok
}

func readFromRegionToMetadataMap(key string) (*PhoneMetadata, bool) {
	v, ok := currentMetadata.regionToMetadataMap[key]
	return v, ok
}

func readFromCountryCodeToNonGeographicalMetadataMap(key int) (*PhoneMetadata, bool) {
	v, ok := currentMetadata.countryCodeToNonGeographicalMetadataMap[key]
	return v, ok
}

var (
	currMetadataColl *PhoneMetadataCollection
	reloadMetadata   = true
)

// MetadataCollection returns the embedded territory metadata collection. The
// result is parsed once and cached; it always reflects the embedded data and
// is independent of any container swapped in via useMetadata.
func MetadataCollection() (*PhoneMetadataCollection, error) {
	if !reloadMetadata {
		return currMetadataColl, nil
	}

	rawBytes, err := decodeUnzip(numberData)
	if err != nil {
		return nil, err
	}

	metadataCollection := &PhoneMetadataCollection{}
	if err = proto.Unmarshal(rawBytes, metadataCollection); err != nil {
		return nil, err
	}
	currMetadataColl = metadataCollection
	reloadMetadata = false
	return metadataCollection, nil
}
