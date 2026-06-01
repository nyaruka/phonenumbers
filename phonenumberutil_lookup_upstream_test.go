package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java, run
// against the synthetic test metadata (see testmetadata_test.go). Method names
// and assertions mirror the Java tests so this file can be kept in sync with
// upstream. Last reconciled against: v9.0.32
//
// This file covers the region/country-code lookup helpers and the
// metadata-loading tests (getMetadataForRegion / getMetadataForNonGeographical
// Region), the Go analogues of the Java testGet* methods in the ranges noted
// inline. Where Java passes a null region code (a Java String null), Go has no
// equivalent for a value-typed string parameter, so those cases use the empty
// string "", which our API treats as an invalid/unknown region.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// testGetSupportedRegions (PhoneNumberUtilTest.java:135-137)
func TestGetSupportedRegions(t *testing.T) {
	useTestMetadata(t)
	assert.Greater(t, len(GetSupportedRegions()), 0)
}

// testGetRegionCodeForCountryCode (PhoneNumberUtilTest.java:1312-1318)
func TestGetRegionCodeForCountryCode(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, regionCode.US, GetRegionCodeForCountryCode(1))
	assert.Equal(t, regionCode.GB, GetRegionCodeForCountryCode(44))
	assert.Equal(t, regionCode.DE, GetRegionCodeForCountryCode(49))
	assert.Equal(t, regionCode.UN001, GetRegionCodeForCountryCode(800))
	assert.Equal(t, regionCode.UN001, GetRegionCodeForCountryCode(979))
}

// testGetRegionCodeForNumber (PhoneNumberUtilTest.java:1320-1326)
func TestGetRegionCodeForNumber(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, regionCode.BS, GetRegionCodeForNumber(bsNumber()))
	assert.Equal(t, regionCode.US, GetRegionCodeForNumber(usNumber()))
	assert.Equal(t, regionCode.GB, GetRegionCodeForNumber(gbMobile()))
	assert.Equal(t, regionCode.UN001, GetRegionCodeForNumber(internationalTollFree()))
	assert.Equal(t, regionCode.UN001, GetRegionCodeForNumber(universalPremiumRate()))
}

// testGetRegionCodesForCountryCode (PhoneNumberUtilTest.java:1328-1337)
func TestGetRegionCodesForCountryCode(t *testing.T) {
	useTestMetadata(t)
	regionCodesForNANPA := GetRegionCodesForCountryCode(1)
	assert.Contains(t, regionCodesForNANPA, regionCode.US)
	assert.Contains(t, regionCodesForNANPA, regionCode.BS)
	assert.Contains(t, GetRegionCodesForCountryCode(44), regionCode.GB)
	assert.Contains(t, GetRegionCodesForCountryCode(49), regionCode.DE)
	assert.Contains(t, GetRegionCodesForCountryCode(800), regionCode.UN001)
	// Test with invalid country calling code.
	assert.Empty(t, GetRegionCodesForCountryCode(-1))
}

// testGetCountryCodeForRegion (PhoneNumberUtilTest.java:1339-1347)
func TestGetCountryCodeForRegion(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, 1, GetCountryCodeForRegion(regionCode.US))
	assert.Equal(t, 64, GetCountryCodeForRegion(regionCode.NZ))
	// Java passes a null region; Go uses the empty string as the no-region case.
	assert.Equal(t, 0, GetCountryCodeForRegion(""))
	assert.Equal(t, 0, GetCountryCodeForRegion(regionCode.ZZ))
	assert.Equal(t, 0, GetCountryCodeForRegion(regionCode.UN001))
	// CS is already deprecated so the library doesn't support it.
	assert.Equal(t, 0, GetCountryCodeForRegion("CS"))
}

// testGetNationalDiallingPrefixForRegion (PhoneNumberUtilTest.java:1349-1364)
func TestGetNationalDiallingPrefixForRegion(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, "1", GetNddPrefixForRegion(regionCode.US, false))
	// Test non-main country to see it gets the national dialling prefix for the
	// main country with that country calling code.
	assert.Equal(t, "1", GetNddPrefixForRegion(regionCode.BS, false))
	assert.Equal(t, "0", GetNddPrefixForRegion(regionCode.NZ, false))
	// Test case with non digit in the national prefix.
	assert.Equal(t, "0~0", GetNddPrefixForRegion(regionCode.AO, false))
	assert.Equal(t, "00", GetNddPrefixForRegion(regionCode.AO, true))
	// Test cases with invalid regions. Java returns null; our API returns the
	// empty string for an invalid/unknown region.
	assert.Equal(t, "", GetNddPrefixForRegion("", false))
	assert.Equal(t, "", GetNddPrefixForRegion(regionCode.ZZ, false))
	assert.Equal(t, "", GetNddPrefixForRegion(regionCode.UN001, false))
	// CS is already deprecated so the library doesn't support it.
	assert.Equal(t, "", GetNddPrefixForRegion("CS", false))
}

// testIsNANPACountry (PhoneNumberUtilTest.java:1366-1373)
func TestIsNANPACountry(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, IsNANPACountry(regionCode.US))
	assert.True(t, IsNANPACountry(regionCode.BS))
	assert.False(t, IsNANPACountry(regionCode.DE))
	assert.False(t, IsNANPACountry(regionCode.ZZ))
	assert.False(t, IsNANPACountry(regionCode.UN001))
	// Java passes a null region; Go uses the empty string as the no-region case.
	assert.False(t, IsNANPACountry(""))
}

// testGetInstanceLoadARMetadata (PhoneNumberUtilTest.java:248-262)
func TestGetInstanceLoadARMetadata(t *testing.T) {
	useTestMetadata(t)
	metadata := getMetadataForRegion(regionCode.AR)
	assert.Equal(t, "AR", metadata.GetId())
	assert.Equal(t, int32(54), metadata.GetCountryCode())
	assert.Equal(t, "00", metadata.GetInternationalPrefix())
	assert.Equal(t, "0", metadata.GetNationalPrefix())
	assert.Equal(t, "0(?:(11|343|3715)15)?", metadata.GetNationalPrefixForParsing())
	assert.Equal(t, "9$1", metadata.GetNationalPrefixTransformRule())
	assert.Equal(t, "$2 15 $3-$4", metadata.GetNumberFormat()[2].GetFormat())
	assert.Equal(t, `(\d)(\d{4})(\d{2})(\d{4})`, metadata.GetNumberFormat()[3].GetPattern())
	assert.Equal(t, `(\d)(\d{4})(\d{2})(\d{4})`, metadata.GetIntlNumberFormat()[3].GetPattern())
	assert.Equal(t, "$1 $2 $3 $4", metadata.GetIntlNumberFormat()[3].GetFormat())
}

// testGetInstanceLoadInternationalTollFreeMetadata (PhoneNumberUtilTest.java:264-273)
func TestGetInstanceLoadInternationalTollFreeMetadata(t *testing.T) {
	useTestMetadata(t)
	metadata := getMetadataForNonGeographicalRegion(800)
	assert.Equal(t, "001", metadata.GetId())
	assert.Equal(t, int32(800), metadata.GetCountryCode())
	assert.Equal(t, "$1 $2", metadata.GetNumberFormat()[0].GetFormat())
	assert.Equal(t, `(\d{4})(\d{4})`, metadata.GetNumberFormat()[0].GetPattern())
	assert.Len(t, metadata.GetGeneralDesc().GetPossibleLengthLocalOnly(), 0)
	assert.Len(t, metadata.GetGeneralDesc().GetPossibleLength(), 1)
	assert.Equal(t, "12345678", metadata.GetTollFree().GetExampleNumber())
}

// testGetMetadataForRegionForNonGeoEntity_shouldBeNull (PhoneNumberUtilTest.java:3249-3251)
func TestGetMetadataForRegionForNonGeoEntityShouldBeNull(t *testing.T) {
	useTestMetadata(t)
	assert.Nil(t, getMetadataForRegion(regionCode.UN001))
}

// testGetMetadataForRegionForUnknownRegion_shouldBeNull (PhoneNumberUtilTest.java:3253-3255)
func TestGetMetadataForRegionForUnknownRegionShouldBeNull(t *testing.T) {
	useTestMetadata(t)
	assert.Nil(t, getMetadataForRegion(regionCode.ZZ))
}

// testGetMetadataForNonGeographicalRegionForGeoRegion_shouldBeNull (PhoneNumberUtilTest.java:3257-3259)
func TestGetMetadataForNonGeographicalRegionForGeoRegionShouldBeNull(t *testing.T) {
	useTestMetadata(t)
	assert.Nil(t, getMetadataForNonGeographicalRegion(1))
}
