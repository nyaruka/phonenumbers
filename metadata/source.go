// Port of metadata/source/* + MetadataLoader from google/libphonenumber.
package metadata

import (
	_ "embed"
	"errors"

	"github.com/nyaruka/phonenumbers/v2/internal/serialize"
	"google.golang.org/protobuf/proto"
)

//go:embed data/countrycode_to_region.xml.gz
var regionData []byte

//go:embed data/metadata.xml.gz
var numberData []byte

// ErrEmptyMetadata is returned when a metadata collection contains no metadata.
var ErrEmptyMetadata = errors.New("empty metadata")

const (
	regionCodeForNonGeoEntity = "001"
	nanpaCountryCode          = 1
)

// Container bundles all of the metadata-derived lookup state used by the
// library. The package keeps a single active container (current), built from
// the embedded metadata during init. Bundling the state behind one swappable
// value is the seam that lets tests run against alternate metadata (e.g.
// upstream's synthetic PhoneNumberMetadataForTesting data) without changing the
// public, package-level API.
type Container struct {
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

// current is the active metadata container. It is populated from the embedded
// metadata by init during package initialization. Tests may swap it via Use to
// exercise the library against alternate metadata.
var current *Container

func init() {
	c, err := Load()
	if err != nil {
		panic(err)
	}
	current = c
}

// Load reads the embedded metadata and builds a fresh container from it without
// changing the active container. It follows the same path used at package init.
func Load() (*Container, error) {
	// load our regions
	regionMap, err := serialize.LoadIntArrayMap(regionData)
	if err != nil {
		return nil, err
	}

	// then our metadata
	coll, err := Collection()
	if err != nil {
		return nil, err
	}

	return NewContainer(coll, regionMap.Map)
}

// NewContainer builds a fully populated Container from a parsed metadata
// collection and a country-code-to-region map. This is the same derivation the
// library has always performed at init, factored out so it can run against any
// metadata source, not just the embedded one.
func NewContainer(coll *PhoneMetadataCollection, countryCodeToRegion map[int][]string) (*Container, error) {
	metadataList := coll.GetMetadata()
	if len(metadataList) == 0 {
		return nil, ErrEmptyMetadata
	}

	mc := &Container{
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
		if region == regionCodeForNonGeoEntity {
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
		if len(regionCodes) == 1 && regionCodeForNonGeoEntity == regionCodes[0] {
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
	delete(mc.supportedRegions, regionCodeForNonGeoEntity)

	for _, val := range countryCodeToRegion[nanpaCountryCode] {
		mc.nanpaRegions[val] = struct{}{}
	}

	return mc, nil
}

// Use swaps the active metadata container, returning a function that restores
// the previously active container. It is intended for tests that need to run
// against alternate (e.g. synthetic) metadata; callers must invoke the returned
// restore function (typically via t.Cleanup) and must not run such tests in
// parallel, since the active container is process-global.
func Use(c *Container) (restore func()) {
	prev := current
	current = c
	return func() { current = prev }
}

// RegionMetadata returns the metadata for the given region code, if supported.
func RegionMetadata(region string) (*PhoneMetadata, bool) {
	v, ok := current.regionToMetadataMap[region]
	return v, ok
}

// NonGeoMetadata returns the metadata for the given non-geographical country
// calling code, if any.
func NonGeoMetadata(countryCode int) (*PhoneMetadata, bool) {
	v, ok := current.countryCodeToNonGeographicalMetadataMap[countryCode]
	return v, ok
}

// IsNANPARegion reports whether region shares country calling code 1.
func IsNANPARegion(region string) bool {
	_, ok := current.nanpaRegions[region]
	return ok
}

// SupportedRegions returns the set of regions the library supports.
func SupportedRegions() map[string]bool { return current.supportedRegions }

// SupportedCallingCodes returns the set of calling codes the library supports.
func SupportedCallingCodes() map[int]bool { return current.supportedCallingCodes }

// CountryCodesForNonGeographicalRegion returns the set of calling codes that map
// to the non-geo entity region ("001").
func CountryCodesForNonGeographicalRegion() map[int]bool {
	return current.countryCodesForNonGeographicalRegion
}

// CountryCodeToRegion returns the map from country calling code to its region
// codes.
func CountryCodeToRegion() map[int][]string { return current.countryCodeToRegion }

var (
	currCollection *PhoneMetadataCollection
	reloadMetadata = true
)

// Collection returns the embedded territory metadata collection. The result is
// parsed once and cached; it always reflects the embedded data and is
// independent of any container swapped in via Use.
func Collection() (*PhoneMetadataCollection, error) {
	if !reloadMetadata {
		return currCollection, nil
	}

	rawBytes, err := serialize.DecodeUnzip(numberData)
	if err != nil {
		return nil, err
	}

	c := &PhoneMetadataCollection{}
	if err = proto.Unmarshal(rawBytes, c); err != nil {
		return nil, err
	}
	currCollection = c
	reloadMetadata = false
	return c, nil
}
