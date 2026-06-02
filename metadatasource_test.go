package phonenumbers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// syntheticCollection builds a tiny, frozen metadata collection for a single
// made-up region so a test can exercise the library against data that never
// changes. This is the Go analogue of upstream's PhoneNumberMetadataForTesting.
func syntheticCollection() (*PhoneMetadataCollection, map[int][]string) {
	sevenDigits := func() *PhoneNumberDesc {
		return &PhoneNumberDesc{
			NationalNumberPattern: proto.String(`\d{7}`),
			PossibleLength:        []int32{7},
		}
	}
	// An absent type is marked with possibleLength [-1] — exactly what the real
	// builder emits for a type with no numbers — so it never matches and
	// getNumberTypeHelper classifies correctly.
	absent := func() *PhoneNumberDesc { return &PhoneNumberDesc{PossibleLength: []int32{-1}} }
	coll := &PhoneMetadataCollection{
		Metadata: []*PhoneMetadata{
			{
				Id:                  proto.String("XX"),
				CountryCode:         proto.Int32(999),
				InternationalPrefix: proto.String("00"),
				GeneralDesc:         sevenDigits(),
				FixedLine:           sevenDigits(),
				Mobile:              absent(),
				PremiumRate:         absent(),
				TollFree:            absent(),
				SharedCost:          absent(),
				Voip:                absent(),
				PersonalNumber:      absent(),
				Pager:               absent(),
				Uan:                 absent(),
				Voicemail:           absent(),
			},
		},
	}
	return coll, map[int][]string{999: {"XX"}}
}

// TestMetadataInjectionSeam proves that useMetadata swaps the data the public,
// package-level API reads from, then cleanly restores the embedded metadata.
// This is the seam that lets us adopt upstream's synthetic-metadata test suite
// without coupling assertions to the real metadata (which changes every release).
func TestMetadataInjectionSeam(t *testing.T) {
	// Baseline: embedded real metadata is active.
	require.NotNil(t, currentMetadata)
	assert.Contains(t, GetSupportedRegions(), "US")
	assert.Equal(t, 1, GetCountryCodeForRegion("US"))
	assert.NotContains(t, GetSupportedRegions(), "XX")

	// Swap in synthetic metadata for a made-up region.
	coll, ccToRegion := syntheticCollection()
	mc, err := newMetadataContainer(coll, ccToRegion)
	require.NoError(t, err)
	restore := useMetadata(mc)
	// Guarantee restoration even if an assertion below panics, so the synthetic
	// metadata can't leak into other tests in the package. restore is
	// idempotent, so the explicit restore() check further down stays valid.
	t.Cleanup(restore)

	// The public API now sees ONLY the synthetic region.
	assert.Equal(t, map[string]bool{"XX": true}, GetSupportedRegions())
	assert.Equal(t, 999, GetCountryCodeForRegion("XX"))
	assert.Equal(t, 0, GetCountryCodeForRegion("US"))
	assert.Equal(t, "XX", GetRegionCodeForCountryCode(999))

	// And it validates/types numbers per the synthetic patterns.
	num := &PhoneNumber{CountryCode: proto.Int32(999), NationalNumber: proto.Uint64(1234567)}
	assert.Equal(t, FIXED_LINE, GetNumberType(num))
	assert.True(t, IsValidNumberForRegion(num, "XX"))

	// Restore and confirm the embedded metadata is back, unchanged.
	restore()
	assert.Contains(t, GetSupportedRegions(), "US")
	assert.NotContains(t, GetSupportedRegions(), "XX")
	assert.Equal(t, 1, GetCountryCodeForRegion("US"))

	// newMetadataContainer rejects an empty collection.
	_, err = newMetadataContainer(&PhoneMetadataCollection{}, nil)
	assert.ErrorIs(t, err, ErrEmptyMetadata)
}
