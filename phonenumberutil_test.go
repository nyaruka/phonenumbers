package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java, run
// against the synthetic test metadata (see testmetadata_test.go). Method names
// and assertions mirror the Java so this file can be kept in sync with upstream.
// Last reconciled against: v9.0.32
//
// This is an in-progress migration of PhoneNumberUtilTest. The remaining ad-hoc,
// real-metadata tests live in phonenumberutil_adhoc_test.go and are deleted from
// there as their faithful ports land here.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func pn(countryCode int32, nationalNumber uint64) *PhoneNumber {
	return &PhoneNumber{CountryCode: proto.Int32(countryCode), NationalNumber: proto.Uint64(nationalNumber)}
}

func alphaNumericNumber() *PhoneNumber { return pn(1, 80074935247) }
func aeUAN() *PhoneNumber              { return pn(971, 600123456) }
func arMobile() *PhoneNumber           { return pn(54, 91187654321) }
func arNumber() *PhoneNumber           { return pn(54, 1187654321) }
func auNumber() *PhoneNumber           { return pn(61, 236618300) }
func bsMobile() *PhoneNumber           { return pn(1, 2423570000) }
func bsNumber() *PhoneNumber           { return pn(1, 2423651234) }
func coFixedLine() *PhoneNumber        { return pn(57, 6012345678) }

// deNumber is the same as the example number for DE in the metadata.
func deNumber() *PhoneNumber      { return pn(49, 30123456) }
func deShortNumber() *PhoneNumber { return pn(49, 1234) }
func gbMobile() *PhoneNumber      { return pn(44, 7912345678) }
func gbNumber() *PhoneNumber      { return pn(44, 2070313000) }
func itMobile() *PhoneNumber      { return pn(39, 345678901) }

func itNumber() *PhoneNumber {
	n := pn(39, 236618300)
	n.ItalianLeadingZero = proto.Bool(true)
	return n
}

func jpStarNumber() *PhoneNumber { return pn(81, 2345) }

// Numbers to test the formatting rules from Mexico.
func mxMobile1() *PhoneNumber { return pn(52, 12345678900) }
func mxMobile2() *PhoneNumber { return pn(52, 15512345678) }
func mxNumber1() *PhoneNumber { return pn(52, 3312345678) }
func mxNumber2() *PhoneNumber { return pn(52, 8211234567) }
func nzNumber() *PhoneNumber  { return pn(64, 33316005) }
func sgNumber() *PhoneNumber  { return pn(65, 65218000) }

// usLongNumber is a too-long and hence invalid US number.
func usLongNumber() *PhoneNumber { return pn(1, 65025300001) }
func usNumber() *PhoneNumber     { return pn(1, 6502530000) }
func usPremium() *PhoneNumber    { return pn(1, 9002530000) }

// Too short, but still possible US numbers.
func usLocalNumber() *PhoneNumber      { return pn(1, 2530000) }
func usShortByOneNumber() *PhoneNumber { return pn(1, 650253000) }
func usTollFree() *PhoneNumber         { return pn(1, 8002530000) }
func usSpoof() *PhoneNumber            { return pn(1, 0) }

func usSpoofWithRawInput() *PhoneNumber {
	n := pn(1, 0)
	n.RawInput = proto.String("000-000-0000")
	return n
}

func uzFixedLine() *PhoneNumber           { return pn(998, 612201234) }
func uzMobile() *PhoneNumber              { return pn(998, 950123456) }
func internationalTollFree() *PhoneNumber { return pn(800, 12345678) }

// internationalTollFreeTooLong is the same length as numbers for the other
// non-geographical country prefix in the test metadata, but is not valid
// because they differ in their country calling code.
func internationalTollFreeTooLong() *PhoneNumber { return pn(800, 123456789) }
func universalPremiumRate() *PhoneNumber         { return pn(979, 123456789) }
func unknownCountryCodeNoRawInput() *PhoneNumber { return pn(2, 12345) }

func TestGetSupportedGlobalNetworkCallingCodes(t *testing.T) {
	useTestMetadata(t)
	globalNetworkCallingCodes := GetSupportedGlobalNetworkCallingCodes()
	assert.NotEmpty(t, globalNetworkCallingCodes)
	for callingCode := range globalNetworkCallingCodes {
		assert.Greater(t, callingCode, 0)
		assert.Equal(t, regionCode.UN001, GetRegionCodeForCountryCode(callingCode))
	}
}

func TestGetSupportedCallingCodes(t *testing.T) {
	useTestMetadata(t)
	callingCodes := GetSupportedCallingCodes()
	assert.NotEmpty(t, callingCodes)
	for callingCode := range callingCodes {
		assert.Greater(t, callingCode, 0)
		assert.NotEqual(t, regionCode.ZZ, GetRegionCodeForCountryCode(callingCode))
	}
	// There should be more than just the global network calling codes in this set.
	assert.Greater(t, len(callingCodes), len(GetSupportedGlobalNetworkCallingCodes()))
	// But they should be included. Testing one of them.
	assert.Contains(t, callingCodes, 979)
}

func TestGetInstanceLoadBadMetadata(t *testing.T) {
	useTestMetadata(t)
	assert.Nil(t, getMetadataForRegion("No Such Region"))
	assert.Nil(t, getMetadataForNonGeographicalRegion(-1))
}

func TestGetInstanceLoadUSMetadata(t *testing.T) {
	useTestMetadata(t)
	metadata := getMetadataForRegion(regionCode.US)
	assert.Equal(t, "US", metadata.GetId())
	assert.Equal(t, int32(1), metadata.GetCountryCode())
	assert.Equal(t, "011", metadata.GetInternationalPrefix())
	assert.NotNil(t, metadata.NationalPrefix) // hasNationalPrefix
	assert.Len(t, metadata.GetNumberFormat(), 2)
	assert.Equal(t, `(\d{3})(\d{3})(\d{4})`, metadata.GetNumberFormat()[1].GetPattern())
	assert.Equal(t, "$1 $2 $3", metadata.GetNumberFormat()[1].GetFormat())
	assert.Equal(t, `[13-689]\d{9}|2[0-35-9]\d{8}`, metadata.GetGeneralDesc().GetNationalNumberPattern())
	assert.Equal(t, `[13-689]\d{9}|2[0-35-9]\d{8}`, metadata.GetFixedLine().GetNationalNumberPattern())
	assert.Len(t, metadata.GetGeneralDesc().GetPossibleLength(), 1)
	assert.Equal(t, int32(10), metadata.GetGeneralDesc().GetPossibleLength()[0])
	// Possible lengths are the same as the general description, so aren't stored
	// separately in the toll free element as well.
	assert.Len(t, metadata.GetTollFree().GetPossibleLength(), 0)
	assert.Equal(t, `900\d{7}`, metadata.GetPremiumRate().GetNationalNumberPattern())
	// No shared-cost data is available for US, so its national number data should
	// not be set (the builder marks the absent descriptor with possibleLength [-1]).
	assert.Nil(t, metadata.GetSharedCost().NationalNumberPattern) // hasNationalNumberPattern() == false
}

func TestGetInstanceLoadDEMetadata(t *testing.T) {
	useTestMetadata(t)
	metadata := getMetadataForRegion(regionCode.DE)
	assert.Equal(t, "DE", metadata.GetId())
	assert.Equal(t, int32(49), metadata.GetCountryCode())
	assert.Equal(t, "00", metadata.GetInternationalPrefix())
	assert.Equal(t, "0", metadata.GetNationalPrefix())
	assert.Len(t, metadata.GetNumberFormat(), 6)
	assert.Len(t, metadata.GetNumberFormat()[5].GetLeadingDigitsPattern(), 1)
	assert.Equal(t, "900", metadata.GetNumberFormat()[5].GetLeadingDigitsPattern()[0])
	assert.Equal(t, `(\d{3})(\d{3,4})(\d{4})`, metadata.GetNumberFormat()[5].GetPattern())
	assert.Equal(t, "$1 $2 $3", metadata.GetNumberFormat()[5].GetFormat())
	assert.Len(t, metadata.GetGeneralDesc().GetPossibleLengthLocalOnly(), 2)
	assert.Len(t, metadata.GetGeneralDesc().GetPossibleLength(), 8)
	// Nothing is present for fixed-line, since it is the same as the general
	// desc, so for efficiency reasons we don't store an extra value.
	assert.Len(t, metadata.GetFixedLine().GetPossibleLength(), 0)
	assert.Len(t, metadata.GetMobile().GetPossibleLength(), 2)
	assert.Equal(t, `(?:[24-6]\d{2}|3[03-9]\d|[789](?:0[2-9]|[1-9]\d))\d{1,8}`, metadata.GetFixedLine().GetNationalNumberPattern())
	assert.Equal(t, "30123456", metadata.GetFixedLine().GetExampleNumber())
	assert.Equal(t, int32(10), metadata.GetTollFree().GetPossibleLength()[0])
	assert.Equal(t, `900([135]\d{6}|9\d{7})`, metadata.GetPremiumRate().GetNationalNumberPattern())
}

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

func TestIsPremiumRate(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, PREMIUM_RATE, GetNumberType(usPremium()))

	premiumRateNumber := &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(892123)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	premiumRateNumber = &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(9187654321)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	premiumRateNumber = &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(9001654321)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	premiumRateNumber = &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(90091234567)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	assert.Equal(t, PREMIUM_RATE, GetNumberType(universalPremiumRate()))
}

func TestIsTollFree(t *testing.T) {
	useTestMetadata(t)

	tollFreeNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(8881234567)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	tollFreeNumber = &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(803123)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	tollFreeNumber = &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(8012345678)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	tollFreeNumber = &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(8001234567)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	assert.Equal(t, TOLL_FREE, GetNumberType(internationalTollFree()))
}

func TestIsMobile(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, MOBILE, GetNumberType(bsMobile()))
	assert.Equal(t, MOBILE, GetNumberType(gbMobile()))
	assert.Equal(t, MOBILE, GetNumberType(itMobile()))
	assert.Equal(t, MOBILE, GetNumberType(arMobile()))

	mobileNumber := &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(15123456789)}
	assert.Equal(t, MOBILE, GetNumberType(mobileNumber))
}

func TestIsFixedLine(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, FIXED_LINE, GetNumberType(bsNumber()))
	assert.Equal(t, FIXED_LINE, GetNumberType(itNumber()))
	assert.Equal(t, FIXED_LINE, GetNumberType(gbNumber()))
	assert.Equal(t, FIXED_LINE, GetNumberType(deNumber()))
}

func TestIsFixedLineAndMobile(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, FIXED_LINE_OR_MOBILE, GetNumberType(usNumber()))

	fixedLineAndMobileNumber := &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(1987654321)}
	assert.Equal(t, FIXED_LINE_OR_MOBILE, GetNumberType(fixedLineAndMobileNumber))
}

func TestIsSharedCost(t *testing.T) {
	useTestMetadata(t)
	gbNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(8431231234)}
	assert.Equal(t, SHARED_COST, GetNumberType(gbNumber))
}

func TestIsVoip(t *testing.T) {
	useTestMetadata(t)
	gbNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(5631231234)}
	assert.Equal(t, VOIP, GetNumberType(gbNumber))
}

func TestIsPersonalNumber(t *testing.T) {
	useTestMetadata(t)
	gbNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(7031231234)}
	assert.Equal(t, PERSONAL_NUMBER, GetNumberType(gbNumber))
}

func TestIsUnknown(t *testing.T) {
	useTestMetadata(t)
	// Invalid numbers should be of type UNKNOWN.
	assert.Equal(t, UNKNOWN, GetNumberType(usLocalNumber()))
}

func TestGetNationalSignificantNumber(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, "6502530000", GetNationalSignificantNumber(usNumber()))

	// An Italian mobile number.
	assert.Equal(t, "345678901", GetNationalSignificantNumber(itMobile()))

	// An Italian fixed line number.
	assert.Equal(t, "0236618300", GetNationalSignificantNumber(itNumber()))

	assert.Equal(t, "12345678", GetNationalSignificantNumber(internationalTollFree()))
}

func TestGetNationalSignificantNumberManyLeadingZeros(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{
		CountryCode:          proto.Int32(1),
		NationalNumber:       proto.Uint64(650),
		ItalianLeadingZero:   proto.Bool(true),
		NumberOfLeadingZeros: proto.Int32(2),
	}
	assert.Equal(t, "00650", GetNationalSignificantNumber(number))

	// Set a bad value; we shouldn't crash, we shouldn't output any leading zeros at all.
	number.NumberOfLeadingZeros = proto.Int32(-3)
	assert.Equal(t, "650", GetNationalSignificantNumber(number))
}

func TestGetLengthOfGeographicalAreaCode(t *testing.T) {
	useTestMetadata(t)
	// Google MTV, which has area code "650".
	assert.Equal(t, 3, GetLengthOfGeographicalAreaCode(usNumber()))

	// A North America toll-free number, which has no area code.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(usTollFree()))

	// Google London, which has area code "20".
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(gbNumber()))

	// A mobile number in the UK does not have an area code (by default, mobile numbers do not,
	// unless they have been added to our list of exceptions).
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(gbMobile()))

	// Google Buenos Aires, which has area code "11".
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(arNumber()))

	// A mobile number in Argentina also has an area code.
	assert.Equal(t, 3, GetLengthOfGeographicalAreaCode(arMobile()))

	// Google Sydney, which has area code "2".
	assert.Equal(t, 1, GetLengthOfGeographicalAreaCode(auNumber()))

	// Italian numbers - there is no national prefix, but it still has an area code.
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(itNumber()))

	// Mexico numbers - there is no national prefix, but it still has an area code.
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(mxNumber1()))

	// Google Singapore. Singapore has no area code and no national prefix.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(sgNumber()))

	// An invalid US number (1 digit shorter), which has no area code.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(usShortByOneNumber()))

	// An international toll free number, which has no area code.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(internationalTollFree()))

	// A mobile number from China is geographical, but does not have an area code.
	cnMobile := &PhoneNumber{CountryCode: proto.Int32(86), NationalNumber: proto.Uint64(18912341234)}
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(cnMobile))
}

func TestGetLengthOfNationalDestinationCode(t *testing.T) {
	useTestMetadata(t)
	// Google MTV, which has national destination code (NDC) "650".
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(usNumber()))

	// A North America toll-free number, which has NDC "800".
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(usTollFree()))

	// Google London, which has NDC "20".
	assert.Equal(t, 2, GetLengthOfNationalDestinationCode(gbNumber()))

	// A UK mobile phone, which has NDC "7912".
	assert.Equal(t, 4, GetLengthOfNationalDestinationCode(gbMobile()))

	// Google Buenos Aires, which has NDC "11".
	assert.Equal(t, 2, GetLengthOfNationalDestinationCode(arNumber()))

	// An Argentinian mobile which has NDC "911".
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(arMobile()))

	// Google Sydney, which has NDC "2".
	assert.Equal(t, 1, GetLengthOfNationalDestinationCode(auNumber()))

	// Google Singapore, which has NDC "6521".
	assert.Equal(t, 4, GetLengthOfNationalDestinationCode(sgNumber()))

	// An invalid US number (1 digit shorter), which has no NDC.
	assert.Equal(t, 0, GetLengthOfNationalDestinationCode(usShortByOneNumber()))

	// A number containing an invalid country calling code, which shouldn't have any NDC.
	number := &PhoneNumber{CountryCode: proto.Int32(123), NationalNumber: proto.Uint64(6502530000)}
	assert.Equal(t, 0, GetLengthOfNationalDestinationCode(number))

	// An international toll free number, which has NDC "1234".
	assert.Equal(t, 4, GetLengthOfNationalDestinationCode(internationalTollFree()))

	// A mobile number from China is geographical, but does not have an area code: however it still
	// can be considered to have a national destination code.
	cnMobile := &PhoneNumber{CountryCode: proto.Int32(86), NationalNumber: proto.Uint64(18912341234)}
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(cnMobile))
}

func TestGetCountryMobileToken(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, "9", GetCountryMobileToken(GetCountryCodeForRegion(regionCode.AR)))

	// Country calling code for Sweden, which has no mobile token.
	assert.Equal(t, "", GetCountryMobileToken(GetCountryCodeForRegion(regionCode.SE)))
}

func TestIsNumberGeographical(t *testing.T) {
	useTestMetadata(t)
	assert.False(t, isNumberGeographical(bsMobile()))              // Bahamas, mobile phone number.
	assert.True(t, isNumberGeographical(auNumber()))               // Australian fixed line number.
	assert.False(t, isNumberGeographical(internationalTollFree())) // International toll free number.
	// We test that mobile phone numbers in relevant regions are indeed considered geographical.
	assert.True(t, isNumberGeographical(arMobile()))  // Argentina, mobile phone number.
	assert.True(t, isNumberGeographical(mxMobile1())) // Mexico, mobile phone number.
	assert.True(t, isNumberGeographical(mxMobile2())) // Mexico, another mobile phone number.
}

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
