package phonenumbers

// Faithful port of PhoneNumberUtilTest.java's validation, example-number and
// misc tests, run against the synthetic test metadata (see testmetadata_test.go).
// Method names and assertions mirror the Java tests so this file can be kept in
// sync with upstream. Last reconciled against: v9.0.32

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestIsValidNumber(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, IsValidNumber(usNumber()))
	assert.True(t, IsValidNumber(itNumber()))
	assert.True(t, IsValidNumber(gbMobile()))
	assert.True(t, IsValidNumber(internationalTollFree()))
	assert.True(t, IsValidNumber(universalPremiumRate()))

	nzNumber := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(21387835)}
	assert.True(t, IsValidNumber(nzNumber))
}

func TestIsValidForRegion(t *testing.T) {
	useTestMetadata(t)
	// This number is valid for the Bahamas, but is not a valid US number.
	assert.True(t, IsValidNumber(bsNumber()))
	assert.True(t, IsValidNumberForRegion(bsNumber(), regionCode.BS))
	assert.False(t, IsValidNumberForRegion(bsNumber(), regionCode.US))
	bsInvalidNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(2421232345)}
	// This number is no longer valid.
	assert.False(t, IsValidNumber(bsInvalidNumber))

	// La Mayotte and Reunion use 'leadingDigits' to differentiate them.
	reNumber := &PhoneNumber{CountryCode: proto.Int32(262), NationalNumber: proto.Uint64(262123456)}
	assert.True(t, IsValidNumber(reNumber))
	assert.True(t, IsValidNumberForRegion(reNumber, regionCode.RE))
	assert.False(t, IsValidNumberForRegion(reNumber, regionCode.YT))
	// Now change the number to be a number for La Mayotte.
	reNumber.NationalNumber = proto.Uint64(269601234)
	assert.True(t, IsValidNumberForRegion(reNumber, regionCode.YT))
	assert.False(t, IsValidNumberForRegion(reNumber, regionCode.RE))
	// This number is no longer valid for La Reunion.
	reNumber.NationalNumber = proto.Uint64(269123456)
	assert.False(t, IsValidNumberForRegion(reNumber, regionCode.YT))
	assert.False(t, IsValidNumberForRegion(reNumber, regionCode.RE))
	assert.False(t, IsValidNumber(reNumber))
	// However, it should be recognised as from La Mayotte, since it is valid for this region.
	assert.Equal(t, regionCode.YT, GetRegionCodeForNumber(reNumber))
	// This number is valid in both places.
	reNumber.NationalNumber = proto.Uint64(800123456)
	assert.True(t, IsValidNumberForRegion(reNumber, regionCode.YT))
	assert.True(t, IsValidNumberForRegion(reNumber, regionCode.RE))
	assert.True(t, IsValidNumberForRegion(internationalTollFree(), regionCode.UN001))
	assert.False(t, IsValidNumberForRegion(internationalTollFree(), regionCode.US))
	assert.False(t, IsValidNumberForRegion(internationalTollFree(), regionCode.ZZ))

	invalidNumber := &PhoneNumber{}
	// Invalid country calling codes.
	invalidNumber.CountryCode = proto.Int32(3923)
	invalidNumber.NationalNumber = proto.Uint64(2366)
	assert.False(t, IsValidNumberForRegion(invalidNumber, regionCode.ZZ))
	assert.False(t, IsValidNumberForRegion(invalidNumber, regionCode.UN001))
	invalidNumber.CountryCode = proto.Int32(0)
	assert.False(t, IsValidNumberForRegion(invalidNumber, regionCode.UN001))
	assert.False(t, IsValidNumberForRegion(invalidNumber, regionCode.ZZ))
}

func TestIsNotValidNumber(t *testing.T) {
	useTestMetadata(t)
	assert.False(t, IsValidNumber(usLocalNumber()))

	invalidNumber := &PhoneNumber{
		CountryCode:        proto.Int32(39),
		NationalNumber:     proto.Uint64(23661830000),
		ItalianLeadingZero: proto.Bool(true),
	}
	assert.False(t, IsValidNumber(invalidNumber))

	invalidNumber = &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(791234567)}
	assert.False(t, IsValidNumber(invalidNumber))

	invalidNumber = &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(1234)}
	assert.False(t, IsValidNumber(invalidNumber))

	invalidNumber = &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(3316005)}
	assert.False(t, IsValidNumber(invalidNumber))

	// Invalid country calling codes.
	invalidNumber = &PhoneNumber{CountryCode: proto.Int32(3923), NationalNumber: proto.Uint64(2366)}
	assert.False(t, IsValidNumber(invalidNumber))
	invalidNumber.CountryCode = proto.Int32(0)
	assert.False(t, IsValidNumber(invalidNumber))

	assert.False(t, IsValidNumber(internationalTollFreeTooLong()))
}

func TestGetExampleNumber(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, proto.Equal(deNumber(), GetExampleNumber(regionCode.DE)))

	assert.True(t, proto.Equal(deNumber(), GetExampleNumberForType(regionCode.DE, FIXED_LINE)))
	// Should return the same response if asked for FIXED_LINE_OR_MOBILE too.
	assert.True(t, proto.Equal(deNumber(), GetExampleNumberForType(regionCode.DE, FIXED_LINE_OR_MOBILE)))
	assert.NotNil(t, GetExampleNumberForType(regionCode.US, FIXED_LINE))
	assert.NotNil(t, GetExampleNumberForType(regionCode.US, MOBILE))

	// We have data for the US, but no data for VOICEMAIL, so return null.
	assert.Nil(t, GetExampleNumberForType(regionCode.US, VOICEMAIL))
	// CS is an invalid region, so we have no data for it.
	assert.Nil(t, GetExampleNumberForType("CS", MOBILE))
	// RegionCode 001 is reserved for supporting non-geographical country calling code. We don't
	// support getting an example number for it with this method.
	assert.Nil(t, GetExampleNumber(regionCode.UN001))
}

func TestGetExampleNumberForNonGeoEntity(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, proto.Equal(internationalTollFree(), GetExampleNumberForNonGeoEntity(800)))
	assert.True(t, proto.Equal(universalPremiumRate(), GetExampleNumberForNonGeoEntity(979)))
}

func TestGetExampleNumberWithoutRegion(t *testing.T) {
	useTestMetadata(t)
	// In our test metadata we don't cover all types: in our real metadata, we do.
	// NOTE: the Go API has no region-less GetExampleNumberForType overload (an
	// upstream API gap, tracked separately), so this mirrors the Java intent by
	// iterating the supported regions and returning the first example for the type.
	assert.NotNil(t, exampleNumberForTypeAnyRegion(FIXED_LINE))
	assert.NotNil(t, exampleNumberForTypeAnyRegion(MOBILE))
	assert.NotNil(t, exampleNumberForTypeAnyRegion(PREMIUM_RATE))
}

// exampleNumberForTypeAnyRegion is the test-local stand-in for upstream's
// getExampleNumberForType(PhoneNumberType) no-region overload: it iterates every
// supported region and returns the first example number found for the given type.
func exampleNumberForTypeAnyRegion(typ PhoneNumberType) *PhoneNumber {
	for regionCode := range GetSupportedRegions() {
		if example := GetExampleNumberForType(regionCode, typ); example != nil {
			return example
		}
	}
	return nil
}

func TestCanBeInternationallyDialled(t *testing.T) {
	useTestMetadata(t)
	// We have no-international-dialling rules for the US in our test metadata that
	// say that toll-free numbers cannot be dialled internationally.
	assert.False(t, canBeInternationallyDialled(usTollFree()))

	// Normal US numbers can be internationally dialled.
	assert.True(t, canBeInternationallyDialled(usNumber()))

	// Invalid number.
	assert.True(t, canBeInternationallyDialled(usLocalNumber()))

	// We have no data for NZ - should return true.
	assert.True(t, canBeInternationallyDialled(nzNumber()))
	assert.True(t, canBeInternationallyDialled(internationalTollFree()))
}

func TestIsAlphaNumber(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, IsAlphaNumber("1800 six-flags"))
	assert.True(t, IsAlphaNumber("1800 six-flags ext. 1234"))
	assert.True(t, IsAlphaNumber("+800 six-flags"))
	assert.True(t, IsAlphaNumber("180 six-flags"))
	assert.False(t, IsAlphaNumber("1800 123-1234"))
	assert.False(t, IsAlphaNumber("1 six-flags"))
	assert.False(t, IsAlphaNumber("18 six-flags"))
	assert.False(t, IsAlphaNumber("1800 123-1234 extension: 1234"))
	assert.False(t, IsAlphaNumber("+800 1234-1234"))
}

func TestIsMobileNumberPortableRegion(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, IsMobileNumberPortableRegion(regionCode.US))
	assert.True(t, IsMobileNumberPortableRegion(regionCode.GB))
	assert.False(t, IsMobileNumberPortableRegion(regionCode.AE))
	assert.False(t, IsMobileNumberPortableRegion(regionCode.BS))
}
