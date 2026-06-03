package phonenumbers

import "github.com/nyaruka/phonenumbers/v2/metadata"

// The core metadata value types are defined in the metadata package (mirroring
// upstream's Phonemetadata.java, and the only acyclic place to put them once the
// metadata loader moves there). They're re-exported here as aliases so the
// public API and the rest of this package keep referring to them unqualified.
type (
	PhoneMetadata           = metadata.PhoneMetadata
	PhoneMetadataCollection = metadata.PhoneMetadataCollection
	PhoneNumberDesc         = metadata.PhoneNumberDesc
	NumberFormat            = metadata.NumberFormat
)
