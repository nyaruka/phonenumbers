// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/metadata/source/* alternate-formats loading from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package phonenumbers

import (
	"sync"

	"github.com/nyaruka/phonenumbers/v2/internal/serialize"
	"google.golang.org/protobuf/proto"
)

// alternateFormatsMap holds the alternate formatting metadata keyed by country
// calling code. The alternate formats provide additional number-format patterns
// (beyond the main metadata) that PhoneNumberMatcher uses for its grouping
// leniency checks. The upstream equivalent is the alternate-formats metadata
// source reached via DefaultMetadataDependenciesProvider.
//
// Loading is deferred to first use (rather than an init) so it never runs before
// the protobuf runtime has finished registering the generated message types.
var (
	alternateFormatsOnce sync.Once
	alternateFormatsMap  = make(map[int32]*PhoneMetadata)
)

func loadAlternateFormatsMetadata() {
	rawBytes, err := serialize.DecodeUnzip(alternateFormatsData)
	if err != nil {
		panic(err)
	}
	if len(rawBytes) == 0 {
		// No alternate-formats data embedded; leave the map empty.
		return
	}
	collection := &PhoneMetadataCollection{}
	if err := proto.Unmarshal(rawBytes, collection); err != nil {
		panic(err)
	}
	for _, meta := range collection.GetMetadata() {
		alternateFormatsMap[meta.GetCountryCode()] = meta
	}
}

// getAlternateFormatsForCountryCallingCode returns the alternate formatting
// metadata for the given country calling code, or nil if none is available.
// Mirrors upstream's getFormattingMetadataForCountryCallingCode.
func getAlternateFormatsForCountryCallingCode(countryCallingCode int32) *PhoneMetadata {
	alternateFormatsOnce.Do(loadAlternateFormatsMetadata)
	return alternateFormatsMap[countryCallingCode]
}
