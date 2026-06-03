package phonenumbers

import "github.com/nyaruka/phonenumbers/v2/metadata"

// The metadata value types from upstream's Phonemetadata. Go's import-cycle rule
// forces their definitions into the metadata package (its loader returns them and
// this package depends on that loader), but upstream keeps them top-level and
// refers to them unqualified — e.g. PhoneNumberUtil's
//
//	PhoneMetadata metadata = getMetadataForRegion(...)
//
// Re-exporting them as aliases keeps this package's references unqualified too,
// so the ported Go stays line-for-line close to the Java source it tracks.
type (
	PhoneMetadata           = metadata.PhoneMetadata
	PhoneMetadataCollection = metadata.PhoneMetadataCollection
	PhoneNumberDesc         = metadata.PhoneNumberDesc
	NumberFormat            = metadata.NumberFormat
)

// MetadataCollection returns the embedded territory metadata collection.
func MetadataCollection() (*PhoneMetadataCollection, error) { return metadata.Collection() }
