package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java, run
// against the synthetic test metadata (see testmetadata_test.go). Method names,
// order, and assertions mirror the Java so this file can be diffed against
// upstream method-by-method. Go-specific internal-helper and regression unit
// tests with no upstream counterpart live in phonenumberutil_internal_test.go.

import (
	"strings"
	"testing"

	"github.com/nyaruka/phonenumbers/v2/internal/stringbuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// testGetSupportedRegions (PhoneNumberUtilTest.java:135-137)

func mustParseAndKeepRawInput(t *testing.T, input, region string) *PhoneNumber {
	t.Helper()
	num, err := ParseAndKeepRawInput(input, region)
	require.NoError(t, err)
	return num
}

// mustParse parses and fails the test on error (see mustParseAndKeepRawInput).
func mustParse(t *testing.T, input, region string) *PhoneNumber {
	t.Helper()
	num, err := Parse(input, region)
	require.NoError(t, err)
	return num
}

// assertThrowsForInvalidPhoneContext mirrors the private Java helper of the same
// name: parsing the number with an unknown region must fail with NOT_A_NUMBER.
func assertThrowsForInvalidPhoneContext(t *testing.T, numberToParse string) {
	_, err := Parse(numberToParse, regionCode.ZZ)
	assert.ErrorIs(t, err, ErrNotANumber, "input %q", numberToParse)
}

// testGetSupportedRegions (PhoneNumberUtilTest.java:135-137)
func TestGetSupportedRegions(t *testing.T) {
	useTestMetadata(t)
	assert.Greater(t, len(GetSupportedRegions()), 0)
}

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

func TestGetSupportedTypesForRegion(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, GetSupportedTypesForRegion(regionCode.BR)[FIXED_LINE])
	// Our test data has no mobile numbers for Brazil.
	assert.False(t, GetSupportedTypesForRegion(regionCode.BR)[MOBILE])
	// UNKNOWN should never be returned.
	assert.False(t, GetSupportedTypesForRegion(regionCode.BR)[UNKNOWN])
	// In the US, many numbers are classified as FIXED_LINE_OR_MOBILE; but we don't want to expose
	// this as a supported type, instead we say FIXED_LINE and MOBILE are both present.
	assert.True(t, GetSupportedTypesForRegion(regionCode.US)[FIXED_LINE])
	assert.True(t, GetSupportedTypesForRegion(regionCode.US)[MOBILE])
	assert.False(t, GetSupportedTypesForRegion(regionCode.US)[FIXED_LINE_OR_MOBILE])

	// Test the invalid region code.
	assert.Equal(t, 0, len(GetSupportedTypesForRegion(regionCode.ZZ)))
}

func TestGetSupportedTypesForNonGeoEntity(t *testing.T) {
	useTestMetadata(t)
	// No data exists for 999 at all, no types should be returned.
	assert.Equal(t, 0, len(GetSupportedTypesForNonGeoEntity(999)))

	typesFor979 := GetSupportedTypesForNonGeoEntity(979)
	assert.True(t, typesFor979[PREMIUM_RATE])
	assert.False(t, typesFor979[MOBILE])
	assert.False(t, typesFor979[UNKNOWN])
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

func TestIsNumberGeographical(t *testing.T) {
	useTestMetadata(t)
	assert.False(t, IsNumberGeographical(bsMobile()))              // Bahamas, mobile phone number.
	assert.True(t, IsNumberGeographical(auNumber()))               // Australian fixed line number.
	assert.False(t, IsNumberGeographical(internationalTollFree())) // International toll free number.
	// We test that mobile phone numbers in relevant regions are indeed considered geographical.
	assert.True(t, IsNumberGeographical(arMobile()))  // Argentina, mobile phone number.
	assert.True(t, IsNumberGeographical(mxMobile1())) // Mexico, mobile phone number.
	assert.True(t, IsNumberGeographical(mxMobile2())) // Mexico, another mobile phone number.
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

func TestGetExampleNumber(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, proto.Equal(deNumber(), GetExampleNumber(regionCode.DE)))

	assert.True(t, proto.Equal(deNumber(), GetExampleNumberForTypeInRegion(regionCode.DE, FIXED_LINE)))
	// Should return the same response if asked for FIXED_LINE_OR_MOBILE too.
	assert.True(t, proto.Equal(deNumber(), GetExampleNumberForTypeInRegion(regionCode.DE, FIXED_LINE_OR_MOBILE)))
	assert.NotNil(t, GetExampleNumberForTypeInRegion(regionCode.US, FIXED_LINE))
	assert.NotNil(t, GetExampleNumberForTypeInRegion(regionCode.US, MOBILE))

	// We have data for the US, but no data for VOICEMAIL, so return null.
	assert.Nil(t, GetExampleNumberForTypeInRegion(regionCode.US, VOICEMAIL))
	// CS is an invalid region, so we have no data for it.
	assert.Nil(t, GetExampleNumberForTypeInRegion("CS", MOBILE))
	// RegionCode 001 is reserved for supporting non-geographical country calling code. We don't
	// support getting an example number for it with this method.
	assert.Nil(t, GetExampleNumber(regionCode.UN001))
}

func TestGetInvalidExampleNumber(t *testing.T) {
	useTestMetadata(t)
	// RegionCode 001 is reserved for supporting non-geographical country calling codes.
	assert.Nil(t, GetInvalidExampleNumber(regionCode.UN001))
	assert.Nil(t, GetInvalidExampleNumber("CS"))
	usInvalidNumber := GetInvalidExampleNumber(regionCode.US)
	assert.Equal(t, int32(1), usInvalidNumber.GetCountryCode())
	assert.NotEqual(t, uint64(0), usInvalidNumber.GetNationalNumber())
}

func TestGetExampleNumberForNonGeoEntity(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, proto.Equal(internationalTollFree(), GetExampleNumberForNonGeoEntity(800)))
	assert.True(t, proto.Equal(universalPremiumRate(), GetExampleNumberForNonGeoEntity(979)))
}

func TestGetExampleNumberWithoutRegion(t *testing.T) {
	useTestMetadata(t)
	// In our test metadata we don't cover all types: in our real metadata, we do.
	assert.NotNil(t, GetExampleNumberForType(FIXED_LINE))
	assert.NotNil(t, GetExampleNumberForType(MOBILE))
	assert.NotNil(t, GetExampleNumberForType(PREMIUM_RATE))
}

// testConvertAlphaCharactersInNumber (PhoneNumberUtilTest.java:446-451)
func TestConvertAlphaCharactersInNumber(t *testing.T) {
	input := "1800-ABC-DEF"
	// Alpha chars are converted to digits; everything else is left untouched.
	assert.Equal(t, "1800-222-333", ConvertAlphaCharactersInNumber(input))
}

// testNormaliseRemovePunctuation (PhoneNumberUtilTest.java:453-458)
func TestNormaliseRemovePunctuation(t *testing.T) {
	assert.Equal(t, "03456234", normalize("034-56&+#2\u00AD34"), "Conversion did not correctly remove punctuation")
}

// testNormaliseReplaceAlphaCharacters (PhoneNumberUtilTest.java:460-465)
func TestNormaliseReplaceAlphaCharacters(t *testing.T) {
	assert.Equal(t, "034426486479", normalize("034-I-am-HUNGRY"), "Conversion did not correctly replace alpha characters")
}

// testNormaliseOtherDigits (PhoneNumberUtilTest.java:467-478)
func TestNormaliseOtherDigits(t *testing.T) {
	assert.Equal(t, "255", normalize("\uFF125\u0665"), "Conversion did not correctly replace non-latin digits")
	// Eastern-Arabic digits.
	assert.Equal(t, "520", normalize("\u06F52\u06F0"), "Conversion did not correctly replace non-latin digits")
}

// testNormaliseStripAlphaCharacters (PhoneNumberUtilTest.java:479-485)
func TestNormaliseStripAlphaCharacters(t *testing.T) {
	assert.Equal(t, "03456234", NormalizeDigitsOnly("034-56&+a#234"), "Conversion did not correctly remove alpha character")
}

// testNormaliseStripNonDiallableCharacters (PhoneNumberUtilTest.java:487-493)
func TestNormaliseStripNonDiallableCharacters(t *testing.T) {
	assert.Equal(t, "03*456+1#234", NormalizeDiallableCharsOnly("03*4-56&+1a#234"), "Conversion did not correctly remove non-diallable characters")
}

func TestFormatUSNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "650 253 0000", Format(usNumber(), NATIONAL))
	assert.Equal(t, "+1 650 253 0000", Format(usNumber(), INTERNATIONAL))

	assert.Equal(t, "800 253 0000", Format(usTollFree(), NATIONAL))
	assert.Equal(t, "+1 800 253 0000", Format(usTollFree(), INTERNATIONAL))

	assert.Equal(t, "900 253 0000", Format(usPremium(), NATIONAL))
	assert.Equal(t, "+1 900 253 0000", Format(usPremium(), INTERNATIONAL))
	assert.Equal(t, "tel:+1-900-253-0000", Format(usPremium(), RFC3966))
	// Numbers with all zeros in the national number part will be formatted by using the raw_input
	// if that is available no matter which format is specified.
	assert.Equal(t, "000-000-0000", Format(usSpoofWithRawInput(), NATIONAL))
	assert.Equal(t, "0", Format(usSpoof(), NATIONAL))
}

func TestFormatBSNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "242 365 1234", Format(bsNumber(), NATIONAL))
	assert.Equal(t, "+1 242 365 1234", Format(bsNumber(), INTERNATIONAL))
}

func TestFormatGBNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "(020) 7031 3000", Format(gbNumber(), NATIONAL))
	assert.Equal(t, "+44 20 7031 3000", Format(gbNumber(), INTERNATIONAL))

	assert.Equal(t, "(07912) 345 678", Format(gbMobile(), NATIONAL))
	assert.Equal(t, "+44 7912 345 678", Format(gbMobile(), INTERNATIONAL))
}

func TestFormatDENumber(t *testing.T) {
	useTestMetadata(t)

	deNumber := pn(49, 301234)
	assert.Equal(t, "030/1234", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 30/1234", Format(deNumber, INTERNATIONAL))
	assert.Equal(t, "tel:+49-30-1234", Format(deNumber, RFC3966))

	deNumber = pn(49, 291123)
	assert.Equal(t, "0291 123", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 291 123", Format(deNumber, INTERNATIONAL))

	deNumber = pn(49, 29112345678)
	assert.Equal(t, "0291 12345678", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 291 12345678", Format(deNumber, INTERNATIONAL))

	deNumber = pn(49, 912312345)
	assert.Equal(t, "09123 12345", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 9123 12345", Format(deNumber, INTERNATIONAL))

	deNumber = pn(49, 80212345)
	assert.Equal(t, "08021 2345", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 8021 2345", Format(deNumber, INTERNATIONAL))
	// Note this number is correctly formatted without national prefix. Most of the numbers that
	// are treated as invalid numbers by the library are short numbers, and they are usually not
	// dialed with national prefix.
	assert.Equal(t, "1234", Format(deShortNumber(), NATIONAL))
	assert.Equal(t, "+49 1234", Format(deShortNumber(), INTERNATIONAL))

	deNumber = pn(49, 41341234)
	assert.Equal(t, "04134 1234", Format(deNumber, NATIONAL))
}

func TestFormatITNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "02 3661 8300", Format(itNumber(), NATIONAL))
	assert.Equal(t, "+39 02 3661 8300", Format(itNumber(), INTERNATIONAL))
	assert.Equal(t, "+390236618300", Format(itNumber(), E164))

	assert.Equal(t, "345 678 901", Format(itMobile(), NATIONAL))
	assert.Equal(t, "+39 345 678 901", Format(itMobile(), INTERNATIONAL))
	assert.Equal(t, "+39345678901", Format(itMobile(), E164))
}

func TestFormatAUNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "02 3661 8300", Format(auNumber(), NATIONAL))
	assert.Equal(t, "+61 2 3661 8300", Format(auNumber(), INTERNATIONAL))
	assert.Equal(t, "+61236618300", Format(auNumber(), E164))

	auNumber := pn(61, 1800123456)
	assert.Equal(t, "1800 123 456", Format(auNumber, NATIONAL))
	assert.Equal(t, "+61 1800 123 456", Format(auNumber, INTERNATIONAL))
	assert.Equal(t, "+611800123456", Format(auNumber, E164))
}

func TestFormatARNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "011 8765-4321", Format(arNumber(), NATIONAL))
	assert.Equal(t, "+54 11 8765-4321", Format(arNumber(), INTERNATIONAL))
	assert.Equal(t, "+541187654321", Format(arNumber(), E164))

	assert.Equal(t, "011 15 8765-4321", Format(arMobile(), NATIONAL))
	assert.Equal(t, "+54 9 11 8765 4321", Format(arMobile(), INTERNATIONAL))
	assert.Equal(t, "+5491187654321", Format(arMobile(), E164))
}

func TestFormatMXNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "045 234 567 8900", Format(mxMobile1(), NATIONAL))
	assert.Equal(t, "+52 1 234 567 8900", Format(mxMobile1(), INTERNATIONAL))
	assert.Equal(t, "+5212345678900", Format(mxMobile1(), E164))

	assert.Equal(t, "045 55 1234 5678", Format(mxMobile2(), NATIONAL))
	assert.Equal(t, "+52 1 55 1234 5678", Format(mxMobile2(), INTERNATIONAL))
	assert.Equal(t, "+5215512345678", Format(mxMobile2(), E164))

	assert.Equal(t, "01 33 1234 5678", Format(mxNumber1(), NATIONAL))
	assert.Equal(t, "+52 33 1234 5678", Format(mxNumber1(), INTERNATIONAL))
	assert.Equal(t, "+523312345678", Format(mxNumber1(), E164))

	assert.Equal(t, "01 821 123 4567", Format(mxNumber2(), NATIONAL))
	assert.Equal(t, "+52 821 123 4567", Format(mxNumber2(), INTERNATIONAL))
	assert.Equal(t, "+528211234567", Format(mxNumber2(), E164))
}

func TestFormatOutOfCountryCallingNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "00 1 900 253 0000", FormatOutOfCountryCallingNumber(usPremium(), regionCode.DE))
	assert.Equal(t, "1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.BS))
	assert.Equal(t, "00 1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.PL))
	assert.Equal(t, "011 44 7912 345 678", FormatOutOfCountryCallingNumber(gbMobile(), regionCode.US))
	assert.Equal(t, "00 49 1234", FormatOutOfCountryCallingNumber(deShortNumber(), regionCode.GB))
	// Note this number is correctly formatted without national prefix.
	assert.Equal(t, "1234", FormatOutOfCountryCallingNumber(deShortNumber(), regionCode.DE))

	assert.Equal(t, "011 39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.US))
	assert.Equal(t, "02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.IT))
	assert.Equal(t, "+39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.SG))

	assert.Equal(t, "6521 8000", FormatOutOfCountryCallingNumber(sgNumber(), regionCode.SG))

	assert.Equal(t, "011 54 9 11 8765 4321", FormatOutOfCountryCallingNumber(arMobile(), regionCode.US))
	assert.Equal(t, "011 800 1234 5678", FormatOutOfCountryCallingNumber(internationalTollFree(), regionCode.US))

	arNumberWithExtn := arMobile()
	arNumberWithExtn.Extension = proto.String("1234")
	assert.Equal(t, "011 54 9 11 8765 4321 ext. 1234", FormatOutOfCountryCallingNumber(arNumberWithExtn, regionCode.US))
	assert.Equal(t, "0011 54 9 11 8765 4321 ext. 1234", FormatOutOfCountryCallingNumber(arNumberWithExtn, regionCode.AU))
	assert.Equal(t, "011 15 8765-4321 ext. 1234", FormatOutOfCountryCallingNumber(arNumberWithExtn, regionCode.AR))
}

func TestFormatOutOfCountryWithInvalidRegion(t *testing.T) {
	useTestMetadata(t)

	// AQ/Antarctica isn't a valid region code for phone number formatting, so this falls back to
	// intl formatting.
	assert.Equal(t, "+1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), "AQ"))
	// For region code 001, the out-of-country format always turns into the international format.
	assert.Equal(t, "+1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.UN001))
}

func TestFormatOutOfCountryWithPreferredIntlPrefix(t *testing.T) {
	useTestMetadata(t)

	// This should use 0011, since that is the preferred international prefix (both 0011 and 0012
	// are accepted as possible international prefixes in our test metadata.)
	assert.Equal(t, "0011 39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.AU))
	// Testing preferred international prefixes with ~ are supported (designates waiting).
	assert.Equal(t, "8~10 39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.UZ))
}

func TestFormatOutOfCountryKeepingAlphaChars(t *testing.T) {
	useTestMetadata(t)

	alphaNumericNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(8007493524), RawInput: proto.String("1800 six-flag")}
	assert.Equal(t, "0011 1 800 SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.AU))

	alphaNumericNumber.RawInput = proto.String("1-800-SIX-flag")
	assert.Equal(t, "0011 1 800-SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.AU))

	alphaNumericNumber.RawInput = proto.String("Call us from UK: 00 1 800 SIX-flag")
	assert.Equal(t, "0011 1 800 SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.AU))

	alphaNumericNumber.RawInput = proto.String("800 SIX-flag")
	assert.Equal(t, "0011 1 800 SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.AU))

	// Formatting from within the NANPA region.
	assert.Equal(t, "1 800 SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.US))
	assert.Equal(t, "1 800 SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.BS))

	// Testing a number with extension.
	alphaNumericNumberWithExtn := mustParseAndKeepRawInput(t, "800 SIX-flag ext. 1234", regionCode.US)
	assert.Equal(t, "0011 1 800 SIX-FLAG extn. 1234", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumberWithExtn, regionCode.AU))

	// Testing that if the raw input doesn't exist, it is formatted using
	// formatOutOfCountryCallingNumber.
	alphaNumericNumber.RawInput = nil
	assert.Equal(t, "00 1 800 749 3524", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.DE))

	// Testing AU alpha number formatted from Australia.
	alphaNumericNumber.CountryCode = proto.Int32(61)
	alphaNumericNumber.NationalNumber = proto.Uint64(827493524)
	alphaNumericNumber.RawInput = proto.String("+61 82749-FLAG")
	// This number should have the national prefix fixed.
	assert.Equal(t, "082749-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.AU))

	alphaNumericNumber.RawInput = proto.String("082749-FLAG")
	assert.Equal(t, "082749-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.AU))

	alphaNumericNumber.NationalNumber = proto.Uint64(18007493524)
	alphaNumericNumber.RawInput = proto.String("1-800-SIX-flag")
	// This number should not have the national prefix prefixed, in accordance with the override for
	// this specific formatting rule.
	assert.Equal(t, "1-800-SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.AU))

	// The metadata should not be permanently changed, since we copied it before modifying patterns.
	alphaNumericNumber.NationalNumber = proto.Uint64(1800749352)
	assert.Equal(t, "1800 749 352", FormatOutOfCountryCallingNumber(alphaNumericNumber, regionCode.AU))

	// Testing a region with multiple international prefixes.
	assert.Equal(t, "+61 1-800-SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.SG))
	// Testing the case of calling from a non-supported region.
	assert.Equal(t, "+61 1-800-SIX-FLAG", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, "AQ"))

	// Testing the case with an invalid country calling code.
	alphaNumericNumber.CountryCode = proto.Int32(0)
	alphaNumericNumber.NationalNumber = proto.Uint64(18007493524)
	alphaNumericNumber.RawInput = proto.String("1-800-SIX-flag")
	// Uses the raw input only.
	assert.Equal(t, "1-800-SIX-flag", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.DE))

	// Testing the case of an invalid alpha number.
	alphaNumericNumber.CountryCode = proto.Int32(1)
	alphaNumericNumber.NationalNumber = proto.Uint64(80749)
	alphaNumericNumber.RawInput = proto.String("180-SIX")
	// No country-code stripping can be done.
	assert.Equal(t, "00 1 180-SIX", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, regionCode.DE))

	// Testing the case of calling from a non-supported region.
	alphaNumericNumber.CountryCode = proto.Int32(1)
	alphaNumericNumber.NationalNumber = proto.Uint64(80749)
	alphaNumericNumber.RawInput = proto.String("180-SIX")
	// No country-code stripping can be done since the number is invalid.
	assert.Equal(t, "+1 180-SIX", FormatOutOfCountryKeepingAlphaChars(alphaNumericNumber, "AQ"))
}

// TestFormatWithCarrierCode is the faithful port of upstream PhoneNumberUtilTest.
// testFormatWithCarrierCode against the synthetic test metadata. AR mobile
// numbers exercise a carrier format whose first $-token is "$2" and whose
// domesticCarrierCodeFormattingRule is "$NP$FG $CC" -> "0$1 $CC"; the "$1" must
// expand to the matched first token ("$2"), not be left literal (which would
// resolve to format group 1 and emit the wrong digit). This guards the
// carrier-code branch of formatNsnUsingPatternWithCarrier.
func TestFormatWithCarrierCode(t *testing.T) {
	useTestMetadata(t)

	arMobile := newPhoneNumber(54, 92234654321)
	assert.Equal(t, "02234 65-4321", Format(arMobile, NATIONAL))
	// Here we force 14 as the carrier code.
	assert.Equal(t, "02234 14 65-4321", FormatNationalNumberWithCarrierCode(arMobile, "14"))
	// Here we force the number to be shown with no carrier code.
	assert.Equal(t, "02234 65-4321", FormatNationalNumberWithCarrierCode(arMobile, ""))
	// E164 ignores national/carrier formatting, so no carrier code is present.
	assert.Equal(t, "+5492234654321", Format(arMobile, E164))
	// We don't support this for the US so there should be no change.
	assert.Equal(t, "650 253 0000", FormatNationalNumberWithCarrierCode(usNumber(), "15"))
	// Invalid country code should just get the NSN.
	assert.Equal(t, "12345", FormatNationalNumberWithCarrierCode(unknownCountryCodeNoRawInput(), "89"))
}

func TestFormatWithPreferredCarrierCode(t *testing.T) {
	useTestMetadata(t)

	// We only support this for AR in our test metadata.
	arNumber := pn(54, 91234125678)
	// Test formatting with no preferred carrier code stored in the number itself.
	assert.Equal(t, "01234 15 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	assert.Equal(t, "01234 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, ""))
	// Test formatting with preferred carrier code present.
	arNumber.PreferredDomesticCarrierCode = proto.String("19")
	assert.Equal(t, "01234 12-5678", Format(arNumber, NATIONAL))
	assert.Equal(t, "01234 19 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	assert.Equal(t, "01234 19 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, ""))
	// When the preferred_domestic_carrier_code is present (even when it is just a space), use it
	// instead of the default carrier code passed in.
	arNumber.PreferredDomesticCarrierCode = proto.String(" ")
	assert.Equal(t, "01234   12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	// When the preferred_domestic_carrier_code is present but empty, treat it as unset and use
	// instead the default carrier code passed in.
	arNumber.PreferredDomesticCarrierCode = proto.String("")
	assert.Equal(t, "01234 15 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	// We don't support this for the US so there should be no change.
	usNumber := pn(1, 4241231234)
	usNumber.PreferredDomesticCarrierCode = proto.String("99")
	assert.Equal(t, "424 123 1234", Format(usNumber, NATIONAL))
	assert.Equal(t, "424 123 1234", FormatNationalNumberWithPreferredCarrierCode(usNumber, "15"))
}

func TestFormatNumberForMobileDialing(t *testing.T) {
	useTestMetadata(t)

	// Numbers are normally dialed in national format in-country, and international format from
	// outside the country.
	assert.Equal(t, "6012345678", FormatNumberForMobileDialing(coFixedLine(), regionCode.CO, false))
	assert.Equal(t, "030123456", FormatNumberForMobileDialing(deNumber(), regionCode.DE, false))
	assert.Equal(t, "+4930123456", FormatNumberForMobileDialing(deNumber(), "CH", false))
	deNumberWithExtn := deNumber()
	deNumberWithExtn.Extension = proto.String("1234")
	assert.Equal(t, "030123456", FormatNumberForMobileDialing(deNumberWithExtn, regionCode.DE, false))
	assert.Equal(t, "+4930123456", FormatNumberForMobileDialing(deNumberWithExtn, "CH", false))

	// US toll free numbers are marked as noInternationalDialling in the test metadata for testing
	// purposes. For such numbers, we expect nothing to be returned when the region code is not the
	// same one.
	assert.Equal(t, "800 253 0000", FormatNumberForMobileDialing(usTollFree(), regionCode.US, true))
	assert.Equal(t, "", FormatNumberForMobileDialing(usTollFree(), regionCode.CN, true))
	assert.Equal(t, "+1 650 253 0000", FormatNumberForMobileDialing(usNumber(), regionCode.US, true))
	usNumberWithExtn := usNumber()
	usNumberWithExtn.Extension = proto.String("1234")
	assert.Equal(t, "+1 650 253 0000", FormatNumberForMobileDialing(usNumberWithExtn, regionCode.US, true))

	assert.Equal(t, "8002530000", FormatNumberForMobileDialing(usTollFree(), regionCode.US, false))
	assert.Equal(t, "", FormatNumberForMobileDialing(usTollFree(), regionCode.CN, false))
	assert.Equal(t, "+16502530000", FormatNumberForMobileDialing(usNumber(), regionCode.US, false))
	assert.Equal(t, "+16502530000", FormatNumberForMobileDialing(usNumberWithExtn, regionCode.US, false))

	// An invalid US number, which is one digit too long.
	assert.Equal(t, "+165025300001", FormatNumberForMobileDialing(usLongNumber(), regionCode.US, false))
	assert.Equal(t, "+1 65025300001", FormatNumberForMobileDialing(usLongNumber(), regionCode.US, true))

	// Star numbers. In real life they appear in Israel, but we have them in JP in our test metadata.
	assert.Equal(t, "*2345", FormatNumberForMobileDialing(jpStarNumber(), regionCode.JP, false))
	assert.Equal(t, "*2345", FormatNumberForMobileDialing(jpStarNumber(), regionCode.JP, true))

	assert.Equal(t, "+80012345678", FormatNumberForMobileDialing(internationalTollFree(), regionCode.JP, false))
	assert.Equal(t, "+800 1234 5678", FormatNumberForMobileDialing(internationalTollFree(), regionCode.JP, true))

	// UAE numbers beginning with 600 (classified as UAN) need to be dialled without +971 locally.
	assert.Equal(t, "+971600123456", FormatNumberForMobileDialing(aeUAN(), regionCode.JP, false))
	assert.Equal(t, "600123456", FormatNumberForMobileDialing(aeUAN(), regionCode.AE, false))

	assert.Equal(t, "+523312345678", FormatNumberForMobileDialing(mxNumber1(), regionCode.MX, false))
	assert.Equal(t, "+523312345678", FormatNumberForMobileDialing(mxNumber1(), regionCode.US, false))

	// Test whether Uzbek phone numbers are returned in international format even when dialled from
	// same region or other regions.
	assert.Equal(t, "+998612201234", FormatNumberForMobileDialing(uzFixedLine(), regionCode.UZ, false))
	assert.Equal(t, "+998950123456", FormatNumberForMobileDialing(uzMobile(), regionCode.UZ, false))
	assert.Equal(t, "+998950123456", FormatNumberForMobileDialing(uzMobile(), regionCode.US, false))

	// Non-geographical numbers should always be dialed in international format.
	assert.Equal(t, "+80012345678", FormatNumberForMobileDialing(internationalTollFree(), regionCode.US, false))
	assert.Equal(t, "+80012345678", FormatNumberForMobileDialing(internationalTollFree(), regionCode.UN001, false))

	// Test that a short number is formatted correctly for mobile dialing within the region, and is
	// not diallable from outside the region.
	deShortNumber := &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(123)}
	assert.Equal(t, "123", FormatNumberForMobileDialing(deShortNumber, regionCode.DE, false))
	assert.Equal(t, "", FormatNumberForMobileDialing(deShortNumber, regionCode.IT, false))

	// Test the special logic for NANPA countries, for which regular length phone numbers are always
	// output in international format, but short numbers are in national format.
	assert.Equal(t, "+16502530000", FormatNumberForMobileDialing(usNumber(), regionCode.US, false))
	assert.Equal(t, "+16502530000", FormatNumberForMobileDialing(usNumber(), regionCode.CA, false))
	assert.Equal(t, "+16502530000", FormatNumberForMobileDialing(usNumber(), regionCode.BR, false))
	usShortNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(911)}
	assert.Equal(t, "911", FormatNumberForMobileDialing(usShortNumber, regionCode.US, false))
	assert.Equal(t, "", FormatNumberForMobileDialing(usShortNumber, regionCode.CA, false))
	assert.Equal(t, "", FormatNumberForMobileDialing(usShortNumber, regionCode.BR, false))

	// Test that the Australian emergency number 000 is formatted correctly.
	auNumber := &PhoneNumber{CountryCode: proto.Int32(61), NationalNumber: proto.Uint64(0), ItalianLeadingZero: proto.Bool(true), NumberOfLeadingZeros: proto.Int32(2)}
	assert.Equal(t, "000", FormatNumberForMobileDialing(auNumber, regionCode.AU, false))
	assert.Equal(t, "", FormatNumberForMobileDialing(auNumber, regionCode.NZ, false))
}

func TestFormatByPattern(t *testing.T) {
	useTestMetadata(t)

	newNumFormat := &NumberFormat{
		Pattern: proto.String("(\\d{3})(\\d{3})(\\d{4})"),
		Format:  proto.String("($1) $2-$3"),
	}
	newNumberFormats := []*NumberFormat{newNumFormat}

	assert.Equal(t, "(650) 253-0000", FormatByPattern(usNumber(), NATIONAL, newNumberFormats))
	assert.Equal(t, "+1 (650) 253-0000", FormatByPattern(usNumber(), INTERNATIONAL, newNumberFormats))
	usNumber2 := pn(1, 6507129823)
	assert.Equal(t, "tel:+1-650-712-9823", FormatByPattern(usNumber2, RFC3966, newNumberFormats))

	// $NP is set to '1' for the US. Here we check that for other NANPA countries the US rules are
	// followed.
	newNumFormat.NationalPrefixFormattingRule = proto.String("$NP ($FG)")
	newNumFormat.Format = proto.String("$1 $2-$3")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "1 (242) 365-1234", FormatByPattern(bsNumber(), NATIONAL, newNumberFormats))
	assert.Equal(t, "+1 242 365-1234", FormatByPattern(bsNumber(), INTERNATIONAL, newNumberFormats))

	newNumFormat.Pattern = proto.String("(\\d{2})(\\d{5})(\\d{3})")
	newNumFormat.Format = proto.String("$1-$2 $3")
	newNumberFormats[0] = newNumFormat

	assert.Equal(t, "02-36618 300", FormatByPattern(itNumber(), NATIONAL, newNumberFormats))
	assert.Equal(t, "+39 02-36618 300", FormatByPattern(itNumber(), INTERNATIONAL, newNumberFormats))

	newNumFormat.NationalPrefixFormattingRule = proto.String("$NP$FG")
	newNumFormat.Pattern = proto.String("(\\d{2})(\\d{4})(\\d{4})")
	newNumFormat.Format = proto.String("$1 $2 $3")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "020 7031 3000", FormatByPattern(gbNumber(), NATIONAL, newNumberFormats))

	newNumFormat.NationalPrefixFormattingRule = proto.String("($NP$FG)")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "(020) 7031 3000", FormatByPattern(gbNumber(), NATIONAL, newNumberFormats))

	newNumFormat.NationalPrefixFormattingRule = proto.String("")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "20 7031 3000", FormatByPattern(gbNumber(), NATIONAL, newNumberFormats))

	assert.Equal(t, "+44 20 7031 3000", FormatByPattern(gbNumber(), INTERNATIONAL, newNumberFormats))
}

// TestFormatByPatternFGProdMetadata is the original bug repro against the real
// embedded metadata: a $NP$FG rule used to emit a literal backslash
// ("0\ 7031 3000") because Go's regexp replacer treated "$1" as a group
// reference into the groupless $FG pattern.
func TestFormatByPatternFGProdMetadata(t *testing.T) {
	num, err := Parse("+442070313000", "GB")
	assert.NoError(t, err)

	formats := []*NumberFormat{{
		Pattern:                      proto.String(`(\d{2})(\d{4})(\d{4})`),
		Format:                       proto.String("$1 $2 $3"),
		NationalPrefixFormattingRule: proto.String("$NP$FG"),
	}}
	assert.Equal(t, "020 7031 3000", FormatByPattern(num, NATIONAL, formats))
}

func TestFormatNationalPrefixFormattingRule(t *testing.T) {
	useTestMetadata(t)

	// AR/MX mobile formats whose first $-token is "$2", with national-prefix
	// rules "0$1"/"$1": the "$1" must expand to the matched token rather than
	// being substituted literally (which would emit format group 1's digit).
	assert.Equal(t, "011 15 8765-4321", Format(arMobile(), NATIONAL))
	assert.Equal(t, "045 234 567 8900", Format(mxMobile1(), NATIONAL))
	assert.Equal(t, "045 55 1234 5678", Format(mxMobile2(), NATIONAL))

	// User-supplied rules containing $FG, compiled by FormatByPattern. $NP is
	// "1" for NANPA regions and "0" for GB.
	bs := &NumberFormat{
		Pattern:                      proto.String(`(\d{3})(\d{3})(\d{4})`),
		Format:                       proto.String("$1 $2-$3"),
		NationalPrefixFormattingRule: proto.String("$NP ($FG)"),
	}
	assert.Equal(t, "1 (242) 365-1234", FormatByPattern(bsNumber(), NATIONAL, []*NumberFormat{bs}))

	gbRule := func(rule string) []*NumberFormat {
		return []*NumberFormat{{
			Pattern:                      proto.String(`(\d{2})(\d{4})(\d{4})`),
			Format:                       proto.String("$1 $2 $3"),
			NationalPrefixFormattingRule: proto.String(rule),
		}}
	}
	assert.Equal(t, "020 7031 3000", FormatByPattern(gbNumber(), NATIONAL, gbRule("$NP$FG")))
	assert.Equal(t, "(020) 7031 3000", FormatByPattern(gbNumber(), NATIONAL, gbRule("($NP$FG)")))
}

func TestFormatE164Number(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "+16502530000", Format(usNumber(), E164))
	assert.Equal(t, "+4930123456", Format(deNumber(), E164))
	assert.Equal(t, "+80012345678", Format(internationalTollFree(), E164))
}

func TestFormatNumberWithExtension(t *testing.T) {
	useTestMetadata(t)

	nzNumber := nzNumber()
	nzNumber.Extension = proto.String("1234")
	// Uses default extension prefix:
	assert.Equal(t, "03-331 6005 ext. 1234", Format(nzNumber, NATIONAL))
	// Uses RFC 3966 syntax.
	assert.Equal(t, "tel:+64-3-331-6005;ext=1234", Format(nzNumber, RFC3966))
	// Extension prefix overridden in the territory information for the US:
	usNumberWithExtension := usNumber()
	usNumberWithExtension.Extension = proto.String("4567")
	assert.Equal(t, "650 253 0000 extn. 4567", Format(usNumberWithExtension, NATIONAL))
}

func TestFormatInOriginalFormat(t *testing.T) {
	useTestMetadata(t)

	number1 := mustParseAndKeepRawInput(t, "+442087654321", regionCode.GB)
	assert.Equal(t, "+44 20 8765 4321", FormatInOriginalFormat(number1, regionCode.GB))

	number2 := mustParseAndKeepRawInput(t, "02087654321", regionCode.GB)
	assert.Equal(t, "(020) 8765 4321", FormatInOriginalFormat(number2, regionCode.GB))

	number3 := mustParseAndKeepRawInput(t, "011442087654321", regionCode.US)
	assert.Equal(t, "011 44 20 8765 4321", FormatInOriginalFormat(number3, regionCode.US))

	number4 := mustParseAndKeepRawInput(t, "442087654321", regionCode.GB)
	assert.Equal(t, "44 20 8765 4321", FormatInOriginalFormat(number4, regionCode.GB))

	number5 := mustParse(t, "+442087654321", regionCode.GB)
	assert.Equal(t, "(020) 8765 4321", FormatInOriginalFormat(number5, regionCode.GB))

	// Invalid numbers that we have a formatting pattern for should be formatted properly. Note area
	// codes starting with 7 are intentionally excluded in the test metadata for testing purposes.
	number6 := mustParseAndKeepRawInput(t, "7345678901", regionCode.US)
	assert.Equal(t, "734 567 8901", FormatInOriginalFormat(number6, regionCode.US))

	// US is not a leading zero country, and the presence of the leading zero leads us to format the
	// number using raw_input.
	number7 := mustParseAndKeepRawInput(t, "0734567 8901", regionCode.US)
	assert.Equal(t, "0734567 8901", FormatInOriginalFormat(number7, regionCode.US))

	// This number is valid, but we don't have a formatting pattern for it. Fall back to the raw
	// input.
	number8 := mustParseAndKeepRawInput(t, "02-4567-8900", regionCode.KR)
	assert.Equal(t, "02-4567-8900", FormatInOriginalFormat(number8, regionCode.KR))

	number9 := mustParseAndKeepRawInput(t, "01180012345678", regionCode.US)
	assert.Equal(t, "011 800 1234 5678", FormatInOriginalFormat(number9, regionCode.US))

	number10 := mustParseAndKeepRawInput(t, "+80012345678", regionCode.KR)
	assert.Equal(t, "+800 1234 5678", FormatInOriginalFormat(number10, regionCode.KR))

	// US local numbers are formatted correctly, as we have formatting patterns for them.
	localNumberUS := mustParseAndKeepRawInput(t, "2530000", regionCode.US)
	assert.Equal(t, "253 0000", FormatInOriginalFormat(localNumberUS, regionCode.US))

	numberWithNationalPrefixUS := mustParseAndKeepRawInput(t, "18003456789", regionCode.US)
	assert.Equal(t, "1 800 345 6789", FormatInOriginalFormat(numberWithNationalPrefixUS, regionCode.US))

	numberWithoutNationalPrefixGB := mustParseAndKeepRawInput(t, "2087654321", regionCode.GB)
	assert.Equal(t, "20 8765 4321", FormatInOriginalFormat(numberWithoutNationalPrefixGB, regionCode.GB))
	// Make sure no metadata is modified as a result of the previous function call.
	assert.Equal(t, "(020) 8765 4321", FormatInOriginalFormat(number5, regionCode.GB))

	numberWithNationalPrefixMX := mustParseAndKeepRawInput(t, "013312345678", regionCode.MX)
	assert.Equal(t, "01 33 1234 5678", FormatInOriginalFormat(numberWithNationalPrefixMX, regionCode.MX))

	numberWithoutNationalPrefixMX := mustParseAndKeepRawInput(t, "3312345678", regionCode.MX)
	assert.Equal(t, "33 1234 5678", FormatInOriginalFormat(numberWithoutNationalPrefixMX, regionCode.MX))

	italianFixedLineNumber := mustParseAndKeepRawInput(t, "0212345678", regionCode.IT)
	assert.Equal(t, "02 1234 5678", FormatInOriginalFormat(italianFixedLineNumber, regionCode.IT))

	numberWithNationalPrefixJP := mustParseAndKeepRawInput(t, "00777012", regionCode.JP)
	assert.Equal(t, "0077-7012", FormatInOriginalFormat(numberWithNationalPrefixJP, regionCode.JP))

	numberWithoutNationalPrefixJP := mustParseAndKeepRawInput(t, "0777012", regionCode.JP)
	assert.Equal(t, "0777012", FormatInOriginalFormat(numberWithoutNationalPrefixJP, regionCode.JP))

	numberWithCarrierCodeBR := mustParseAndKeepRawInput(t, "012 3121286979", regionCode.BR)
	assert.Equal(t, "012 3121286979", FormatInOriginalFormat(numberWithCarrierCodeBR, regionCode.BR))

	// The default national prefix used in this case is 045. When a number with national prefix 044
	// is entered, we return the raw input as we don't want to change the number entered.
	numberWithNationalPrefixMX1 := mustParseAndKeepRawInput(t, "044(33)1234-5678", regionCode.MX)
	assert.Equal(t, "044(33)1234-5678", FormatInOriginalFormat(numberWithNationalPrefixMX1, regionCode.MX))

	numberWithNationalPrefixMX2 := mustParseAndKeepRawInput(t, "045(33)1234-5678", regionCode.MX)
	assert.Equal(t, "045 33 1234 5678", FormatInOriginalFormat(numberWithNationalPrefixMX2, regionCode.MX))

	// The default international prefix used in this case is 0011. When a number with international
	// prefix 0012 is entered, we return the raw input as we don't want to change the number entered.
	outOfCountryNumberFromAU1 := mustParseAndKeepRawInput(t, "0012 16502530000", regionCode.AU)
	assert.Equal(t, "0012 16502530000", FormatInOriginalFormat(outOfCountryNumberFromAU1, regionCode.AU))

	outOfCountryNumberFromAU2 := mustParseAndKeepRawInput(t, "0011 16502530000", regionCode.AU)
	assert.Equal(t, "0011 1 650 253 0000", FormatInOriginalFormat(outOfCountryNumberFromAU2, regionCode.AU))

	// Test the star sign is not removed from or added to the original input by this method.
	starNumber := mustParseAndKeepRawInput(t, "*1234", regionCode.JP)
	assert.Equal(t, "*1234", FormatInOriginalFormat(starNumber, regionCode.JP))
	numberWithoutStar := mustParseAndKeepRawInput(t, "1234", regionCode.JP)
	assert.Equal(t, "1234", FormatInOriginalFormat(numberWithoutStar, regionCode.JP))

	// Test an invalid national number without raw input is just formatted as the national number.
	assert.Equal(t, "650253000", FormatInOriginalFormat(usShortByOneNumber(), regionCode.US))
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

// testIsPossibleNumber (PhoneNumberUtilTest.java:1375-1391)
func TestIsPossibleNumber(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, IsPossibleNumber(usNumber()))
	assert.True(t, IsPossibleNumber(usLocalNumber()))
	assert.True(t, IsPossibleNumber(gbNumber()))
	assert.True(t, IsPossibleNumber(internationalTollFree()))

	assert.True(t, IsPossibleNumberFromRegion("+1 650 253 0000", regionCode.US))
	assert.True(t, IsPossibleNumberFromRegion("+1 650 GOO OGLE", regionCode.US))
	assert.True(t, IsPossibleNumberFromRegion("(650) 253-0000", regionCode.US))
	assert.True(t, IsPossibleNumberFromRegion("253-0000", regionCode.US))
	assert.True(t, IsPossibleNumberFromRegion("+1 650 253 0000", regionCode.GB))
	assert.True(t, IsPossibleNumberFromRegion("+44 20 7031 3000", regionCode.GB))
	assert.True(t, IsPossibleNumberFromRegion("(020) 7031 300", regionCode.GB))
	assert.True(t, IsPossibleNumberFromRegion("7031 3000", regionCode.GB))
	assert.True(t, IsPossibleNumberFromRegion("3331 6005", regionCode.NZ))
	assert.True(t, IsPossibleNumberFromRegion("+800 1234 5678", regionCode.UN001))
}

func TestIsPossibleNumberForTypeDifferentTypeLengths(t *testing.T) {
	useTestMetadata(t)
	// We use Argentinian numbers since they have different possible lengths for different types.
	number := &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(12345)}
	// Too short for any Argentinian number, including fixed-line.
	assert.False(t, IsPossibleNumberForType(number, FIXED_LINE))
	assert.False(t, IsPossibleNumberForType(number, UNKNOWN))

	// 6-digit numbers are okay for fixed-line.
	number.NationalNumber = proto.Uint64(123456)
	assert.True(t, IsPossibleNumberForType(number, UNKNOWN))
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE))
	// But too short for mobile.
	assert.False(t, IsPossibleNumberForType(number, MOBILE))
	// And too short for toll-free.
	assert.False(t, IsPossibleNumberForType(number, TOLL_FREE))

	// The same applies to 9-digit numbers.
	number.NationalNumber = proto.Uint64(123456789)
	assert.True(t, IsPossibleNumberForType(number, UNKNOWN))
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE))
	assert.False(t, IsPossibleNumberForType(number, MOBILE))
	assert.False(t, IsPossibleNumberForType(number, TOLL_FREE))

	// 10-digit numbers are universally possible.
	number.NationalNumber = proto.Uint64(1234567890)
	assert.True(t, IsPossibleNumberForType(number, UNKNOWN))
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE))
	assert.True(t, IsPossibleNumberForType(number, MOBILE))
	assert.True(t, IsPossibleNumberForType(number, TOLL_FREE))

	// 11-digit numbers are only possible for mobile numbers.
	number.NationalNumber = proto.Uint64(12345678901)
	assert.True(t, IsPossibleNumberForType(number, UNKNOWN))
	assert.False(t, IsPossibleNumberForType(number, FIXED_LINE))
	assert.True(t, IsPossibleNumberForType(number, MOBILE))
	assert.False(t, IsPossibleNumberForType(number, TOLL_FREE))
}

func TestIsPossibleNumberForTypeLocalOnly(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(12)}
	// Here we test a number length which matches a local-only length.
	assert.True(t, IsPossibleNumberForType(number, UNKNOWN))
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE))
	// Mobile numbers must be 10 or 11 digits, and there are no local-only lengths.
	assert.False(t, IsPossibleNumberForType(number, MOBILE))
}

func TestIsPossibleNumberForTypeDataMissingForSizeReasons(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{CountryCode: proto.Int32(55), NationalNumber: proto.Uint64(12345678)}
	// Local-only number.
	assert.True(t, IsPossibleNumberForType(number, UNKNOWN))
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE))

	number.NationalNumber = proto.Uint64(1234567890)
	assert.True(t, IsPossibleNumberForType(number, UNKNOWN))
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE))
}

func TestIsPossibleNumberForTypeNumberTypeNotSupportedForRegion(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{CountryCode: proto.Int32(55), NationalNumber: proto.Uint64(12345678)}
	// There are *no* mobile numbers for this region at all, so we return false.
	assert.False(t, IsPossibleNumberForType(number, MOBILE))
	// This matches a fixed-line length though.
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE))
	assert.True(t, IsPossibleNumberForType(number, FIXED_LINE_OR_MOBILE))

	// There are *no* fixed-line OR mobile numbers for this country calling code at all.
	number = &PhoneNumber{CountryCode: proto.Int32(979), NationalNumber: proto.Uint64(123456789)}
	assert.False(t, IsPossibleNumberForType(number, MOBILE))
	assert.False(t, IsPossibleNumberForType(number, FIXED_LINE))
	assert.False(t, IsPossibleNumberForType(number, FIXED_LINE_OR_MOBILE))
	assert.True(t, IsPossibleNumberForType(number, PREMIUM_RATE))
}

// testIsPossibleNumberWithReason (PhoneNumberUtilTest.java:1475-1500)
func TestIsPossibleNumberWithReason(t *testing.T) {
	useTestMetadata(t)
	// National numbers for country calling code +1 that are within 7 to 10 digits are possible.
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberWithReason(usNumber()))

	assert.Equal(t, IS_POSSIBLE_LOCAL_ONLY, IsPossibleNumberWithReason(usLocalNumber()))

	assert.Equal(t, TOO_LONG, IsPossibleNumberWithReason(usLongNumber()))

	number := &PhoneNumber{CountryCode: proto.Int32(0), NationalNumber: proto.Uint64(2530000)}
	assert.Equal(t, INVALID_COUNTRY_CODE, IsPossibleNumberWithReason(number))

	number = &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(253000)}
	assert.Equal(t, TOO_SHORT, IsPossibleNumberWithReason(number))

	number = &PhoneNumber{CountryCode: proto.Int32(65), NationalNumber: proto.Uint64(1234567890)}
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberWithReason(number))

	assert.Equal(t, TOO_LONG, IsPossibleNumberWithReason(internationalTollFreeTooLong()))
}

func TestIsPossibleNumberForTypeWithReasonDifferentTypeLengths(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(12345)}
	// Too short for any Argentinian number.
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))

	// 6-digit numbers are okay for fixed-line.
	number.NationalNumber = proto.Uint64(123456)
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, TOLL_FREE))

	// The same applies to 9-digit numbers.
	number.NationalNumber = proto.Uint64(123456789)
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, TOLL_FREE))

	// 10-digit numbers are universally possible.
	number.NationalNumber = proto.Uint64(1234567890)
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, TOLL_FREE))

	// 11-digit numbers are only possible for mobile numbers.
	number.NationalNumber = proto.Uint64(12345678901)
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, TOLL_FREE))
}

func TestIsPossibleNumberForTypeWithReasonLocalOnly(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(12)}
	assert.Equal(t, IS_POSSIBLE_LOCAL_ONLY, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, IS_POSSIBLE_LOCAL_ONLY, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, MOBILE))
}

func TestIsPossibleNumberForTypeWithReasonDataMissingForSizeReasons(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{CountryCode: proto.Int32(55), NationalNumber: proto.Uint64(12345678)}
	assert.Equal(t, IS_POSSIBLE_LOCAL_ONLY, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, IS_POSSIBLE_LOCAL_ONLY, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))

	number.NationalNumber = proto.Uint64(1234567890)
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, UNKNOWN))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
}

func TestIsPossibleNumberForTypeWithReasonNumberTypeNotSupportedForRegion(t *testing.T) {
	useTestMetadata(t)
	// There are *no* mobile numbers for this region at all, so we return INVALID_LENGTH.
	number := &PhoneNumber{CountryCode: proto.Int32(55), NationalNumber: proto.Uint64(12345678)}
	assert.Equal(t, INVALID_LENGTH, IsPossibleNumberForTypeWithReason(number, MOBILE))
	// This matches a fixed-line length though.
	assert.Equal(t, IS_POSSIBLE_LOCAL_ONLY, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))
	// This is too short for fixed-line, and no mobile numbers exist.
	number = &PhoneNumber{CountryCode: proto.Int32(55), NationalNumber: proto.Uint64(1234567)}
	assert.Equal(t, INVALID_LENGTH, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))

	// This is too short for mobile, and no fixed-line numbers exist.
	number = &PhoneNumber{CountryCode: proto.Int32(882), NationalNumber: proto.Uint64(1234567)}
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))
	assert.Equal(t, INVALID_LENGTH, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))

	// There are *no* fixed-line OR mobile numbers for this country calling code at all.
	number = &PhoneNumber{CountryCode: proto.Int32(979), NationalNumber: proto.Uint64(123456789)}
	assert.Equal(t, INVALID_LENGTH, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, INVALID_LENGTH, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, INVALID_LENGTH, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, PREMIUM_RATE))
}

func TestIsPossibleNumberForTypeWithReasonFixedLineOrMobile(t *testing.T) {
	useTestMetadata(t)
	// For FIXED_LINE_OR_MOBILE, a number should be considered valid if it matches the possible
	// lengths for mobile *or* fixed-line numbers.
	number := &PhoneNumber{CountryCode: proto.Int32(290), NationalNumber: proto.Uint64(1234)}
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))

	number.NationalNumber = proto.Uint64(12345)
	assert.Equal(t, TOO_SHORT, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, INVALID_LENGTH, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))

	number.NationalNumber = proto.Uint64(123456)
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))

	number.NationalNumber = proto.Uint64(1234567)
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, FIXED_LINE))
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, MOBILE))
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))

	// 8-digit numbers are possible for toll-free and premium-rate numbers only.
	number.NationalNumber = proto.Uint64(12345678)
	assert.Equal(t, IS_POSSIBLE, IsPossibleNumberForTypeWithReason(number, TOLL_FREE))
	assert.Equal(t, TOO_LONG, IsPossibleNumberForTypeWithReason(number, FIXED_LINE_OR_MOBILE))
}

// testIsNotPossibleNumber (PhoneNumberUtilTest.java:1727-1745)
func TestIsNotPossibleNumber(t *testing.T) {
	useTestMetadata(t)
	assert.False(t, IsPossibleNumber(usLongNumber()))
	assert.False(t, IsPossibleNumber(internationalTollFreeTooLong()))

	number := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(253000)}
	assert.False(t, IsPossibleNumber(number))

	number = &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(300)}
	assert.False(t, IsPossibleNumber(number))
	assert.False(t, IsPossibleNumberFromRegion("+1 650 253 00000", regionCode.US))
	assert.False(t, IsPossibleNumberFromRegion("(650) 253-00000", regionCode.US))
	assert.False(t, IsPossibleNumberFromRegion("I want a Pizza", regionCode.US))
	assert.False(t, IsPossibleNumberFromRegion("253-000", regionCode.US))
	assert.False(t, IsPossibleNumberFromRegion("1 3000", regionCode.GB))
	assert.False(t, IsPossibleNumberFromRegion("+44 300", regionCode.GB))
	assert.False(t, IsPossibleNumberFromRegion("+800 1234 5678 9", regionCode.UN001))
}

// testTruncateTooLongNumber (PhoneNumberUtilTest.java:1747-1796)
func TestTruncateTooLongNumber(t *testing.T) {
	useTestMetadata(t)
	// GB number 080 1234 5678, but entered with 4 extra digits at the end.
	tooLongNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(80123456780123)}
	validNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(8012345678)}
	assert.True(t, TruncateTooLongNumber(tooLongNumber))
	assert.True(t, proto.Equal(validNumber, tooLongNumber))

	// IT number 022 3456 7890, but entered with 3 extra digits at the end.
	tooLongNumber = &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(2234567890123), ItalianLeadingZero: proto.Bool(true)}
	validNumber = &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(2234567890), ItalianLeadingZero: proto.Bool(true)}
	assert.True(t, TruncateTooLongNumber(tooLongNumber))
	assert.True(t, proto.Equal(validNumber, tooLongNumber))

	// US number 650-253-0000, but entered with one additional digit at the end.
	tooLongNumber = usLongNumber()
	assert.True(t, TruncateTooLongNumber(tooLongNumber))
	assert.True(t, proto.Equal(usNumber(), tooLongNumber))

	tooLongNumber = internationalTollFreeTooLong()
	assert.True(t, TruncateTooLongNumber(tooLongNumber))
	assert.True(t, proto.Equal(internationalTollFree(), tooLongNumber))

	// Tests what happens when a valid number is passed in.
	validNumberCopy := proto.Clone(validNumber).(*PhoneNumber)
	assert.True(t, TruncateTooLongNumber(validNumber))
	// Tests the number is not modified.
	assert.True(t, proto.Equal(validNumberCopy, validNumber))

	// Tests what happens when a number with invalid prefix is passed in.
	// The test metadata says US numbers cannot have prefix 240.
	numberWithInvalidPrefix := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(2401234567)}
	invalidNumberCopy := proto.Clone(numberWithInvalidPrefix).(*PhoneNumber)
	assert.False(t, TruncateTooLongNumber(numberWithInvalidPrefix))
	// Tests the number is not modified.
	assert.True(t, proto.Equal(invalidNumberCopy, numberWithInvalidPrefix))

	// Tests what happens when a too short number is passed in.
	tooShortNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(1234)}
	tooShortNumberCopy := proto.Clone(tooShortNumber).(*PhoneNumber)
	assert.False(t, TruncateTooLongNumber(tooShortNumber))
	// Tests the number is not modified.
	assert.True(t, proto.Equal(tooShortNumberCopy, tooShortNumber))
}

// testIsViablePhoneNumber (PhoneNumberUtilTest.java:1798-1813)
func TestIsViablePhoneNumber(t *testing.T) {
	assert.False(t, isViablePhoneNumber("1"))
	// Only one or two digits before strange non-possible punctuation.
	assert.False(t, isViablePhoneNumber("1+1+1"))
	assert.False(t, isViablePhoneNumber("80+0"))
	// Two digits is viable.
	assert.True(t, isViablePhoneNumber("00"))
	assert.True(t, isViablePhoneNumber("111"))
	// Alpha numbers.
	assert.True(t, isViablePhoneNumber("0800-4-pizza"))
	assert.True(t, isViablePhoneNumber("0800-4-PIZZA"))
	// We need at least three digits before any alpha characters.
	assert.False(t, isViablePhoneNumber("08-PIZZA"))
	assert.False(t, isViablePhoneNumber("8-PIZZA"))
	assert.False(t, isViablePhoneNumber("12. March"))
}

// testIsViablePhoneNumberNonAscii (PhoneNumberUtilTest.java:1815-1823)
func TestIsViablePhoneNumberNonAscii(t *testing.T) {
	// Only one or two digits before possible punctuation followed by more digits.
	assert.True(t, isViablePhoneNumber("1\u300034"))
	assert.False(t, isViablePhoneNumber("1\u30003+4"))
	// Unicode variants of possible starting character and other allowed punctuation/digits.
	assert.True(t, isViablePhoneNumber("\uFF081\uFF09\u30003456789"))
	// Testing a leading + is okay.
	assert.True(t, isViablePhoneNumber("+1\uFF09\u30003456789"))
}

// testExtractPossibleNumber (PhoneNumberUtilTest.java:1825-1847)
func TestExtractPossibleNumber(t *testing.T) {
	// Removes preceding funky punctuation and letters but leaves the rest untouched.
	assert.Equal(t, "0800-345-600", extractPossibleNumber("Tel:0800-345-600"))
	assert.Equal(t, "0800 FOR PIZZA", extractPossibleNumber("Tel:0800 FOR PIZZA"))
	// Should not remove plus sign
	assert.Equal(t, "+800-345-600", extractPossibleNumber("Tel:+800-345-600"))
	// Should recognise wide digits as possible start values.
	assert.Equal(t, "\uFF10\uFF12\uFF13", extractPossibleNumber("\uFF10\uFF12\uFF13"))
	// Dashes are not possible start values and should be removed.
	assert.Equal(t, "\uFF11\uFF12\uFF13", extractPossibleNumber("Num-\uFF11\uFF12\uFF13"))
	// If not possible number present, return empty string.
	assert.Equal(t, "", extractPossibleNumber("Num-...."))
	// Leading brackets are stripped - these are not used when parsing.
	assert.Equal(t, "650) 253-0000", extractPossibleNumber("(650) 253-0000"))

	// Trailing non-alpha-numeric characters should be removed.
	assert.Equal(t, "650) 253-0000", extractPossibleNumber("(650) 253-0000..- .."))
	assert.Equal(t, "650) 253-0000", extractPossibleNumber("(650) 253-0000."))
	// This case has a trailing RTL char.
	assert.Equal(t, "650) 253-0000", extractPossibleNumber("(650) 253-0000\u200F"))
}

// TestMaybeStripNationalPrefixAndCarrierCode is the faithful port of upstream
// PhoneNumberUtilTest.testMaybeStripNationalPrefix. It builds its own metadata,
// so it doesn't need useTestMetadata.
func TestMaybeStripNationalPrefixAndCarrierCode(t *testing.T) {
	// Test basic national prefix stripping
	metadata := &PhoneMetadata{}
	metadata.NationalPrefixForParsing = proto.String("34")
	metadata.GeneralDesc = &PhoneNumberDesc{NationalNumberPattern: proto.String("\\d{4,8}")}

	number := stringbuilder.New([]byte("34356778"))
	assert.True(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "356778", number.String())

	// Retry - should not strip again
	assert.False(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "356778", number.String())

	// No national prefix
	metadata.NationalPrefixForParsing = proto.String("")
	number = stringbuilder.New([]byte("356778"))
	assert.False(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "356778", number.String())

	// If stripping doesn't match national rule, don't strip
	metadata.NationalPrefixForParsing = proto.String("3")
	number = stringbuilder.New([]byte("3123"))
	assert.False(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "3123", number.String())

	// Test extracting carrier code
	metadata.NationalPrefixForParsing = proto.String("0(81)?")
	number = stringbuilder.New([]byte("08122123456"))
	carrierCode := stringbuilder.New(nil)
	assert.True(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, carrierCode))
	assert.Equal(t, "81", carrierCode.String())
	assert.Equal(t, "22123456", number.String())

	// Test with transform rule
	metadata.NationalPrefixTransformRule = proto.String("5${1}5")
	metadata.NationalPrefixForParsing = proto.String("0(\\d{2})")
	number = stringbuilder.New([]byte("031123"))
	assert.True(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "5315123", number.String())
}

// testMaybeStripInternationalPrefix (PhoneNumberUtilTest.java:1902-1960)
func TestMaybeStripInternationalPrefix(t *testing.T) {
	internationalPrefix := "00[39]"
	numberToStrip := stringbuilder.New([]byte("0034567700-3898003"))
	// Note the dash is removed as part of the normalization.
	strippedNumber := "45677003898003"
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of its international prefix.")
	// Now the number no longer starts with an IDD prefix, so it should now report
	// FROM_DEFAULT_COUNTRY.
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))

	numberToStrip = stringbuilder.New([]byte("00945677003898003"))
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of its international prefix.")
	// Test it works when the international prefix is broken up by spaces.
	numberToStrip = stringbuilder.New([]byte("00 9 45677003898003"))
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of its international prefix.")
	// Now the number no longer starts with an IDD prefix, so it should now report
	// FROM_DEFAULT_COUNTRY.
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))

	// Test the + symbol is also recognised and stripped.
	numberToStrip = stringbuilder.New([]byte("+45677003898003"))
	strippedNumber = "45677003898003"
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of the plus symbol.")

	// If the number afterwards is a zero, we should not strip this - no country calling code begins
	// with 0.
	numberToStrip = stringbuilder.New([]byte("0090112-3123"))
	strippedNumber = "00901123123"
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied had a 0 after the match so shouldn't be stripped.")
	// Here the 0 is separated by a space from the IDD.
	numberToStrip = stringbuilder.New([]byte("009 0-112-3123"))
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
}

// testMaybeExtractCountryCode (PhoneNumberUtilTest.java:1962-2090)
func TestMaybeExtractCountryCode(t *testing.T) {
	useTestMetadata(t)
	metadata := getMetadataForRegion(regionCode.US)

	// Note that for the US, the IDD is 011.
	number := &PhoneNumber{}
	numberToFill := stringbuilder.New(nil)
	cc, err := maybeExtractCountryCode("011112-3456789", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 1, cc, "Did not extract country calling code 1 correctly.")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")
	// Should strip and normalize national significant number.
	assert.Equal(t, "123456789", numberToFill.String(), "Did not strip off the country calling code correctly.")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	cc, err = maybeExtractCountryCode("+6423456789", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 64, cc, "Did not extract country calling code 64 correctly.")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	cc, err = maybeExtractCountryCode("+80012345678", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 800, cc, "Did not extract country calling code 800 correctly.")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	cc, err = maybeExtractCountryCode("2345-6789", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 0, cc, "Should not have extracted a country calling code - no international prefix present.")
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	_, err = maybeExtractCountryCode("0119991123456789", metadata, numberToFill, true, number)
	assert.ErrorIs(t, err, ErrInvalidCountryCode, "Should have thrown an exception, no valid country calling code present.")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	cc, err = maybeExtractCountryCode("(1 610) 619 4466", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 1, cc, "Should have extracted the country calling code of the region passed in")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITHOUT_PLUS_SIGN, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	cc, err = maybeExtractCountryCode("(1 610) 619 4466", metadata, numberToFill, false, number)
	assert.NoError(t, err)
	assert.Equal(t, 1, cc, "Should have extracted the country calling code of the region passed in")
	assert.Nil(t, number.CountryCodeSource, "Should not contain CountryCodeSource.")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	cc, err = maybeExtractCountryCode("(1 610) 619 446", metadata, numberToFill, false, number)
	assert.NoError(t, err)
	assert.Equal(t, 0, cc, "Should not have extracted a country calling code - invalid number after extraction of uncertain country calling code.")
	assert.Nil(t, number.CountryCodeSource, "Should not contain CountryCodeSource.")

	number = &PhoneNumber{}
	numberToFill = stringbuilder.New(nil)
	cc, err = maybeExtractCountryCode("(1 610) 619", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 0, cc, "Should not have extracted a country calling code - too short number both before and after extraction of uncertain country calling code.")
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")
}

func TestParseNationalNumber(t *testing.T) {
	useTestMetadata(t)

	// National prefix attached.
	got, err := Parse("033316005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Some fields are not filled in by parse, but only by parseAndKeepRawInput.
	// (Java also asserts hasCountryCodeSource() is false; our proto has no
	// presence distinct from the zero/UNSPECIFIED value, so this single check
	// covers both Java assertions.)
	assert.Equal(t, PhoneNumber_UNSPECIFIED, got.GetCountryCodeSource())

	got, err = Parse("33316005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// National prefix attached and some formatting present.
	got, err = Parse("03-331 6005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("03 331 6005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Test parsing RFC3966 format with a phone context.
	got, err = Parse("tel:03-331-6005;phone-context=+64", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("tel:331-6005;phone-context=+64-3", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("tel:331-6005;phone-context=+64-3", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("My number is tel:03-331-6005;phone-context=+64", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Test parsing RFC3966 format with optional user-defined parameters. The parameters will appear
	// after the context if present.
	got, err = Parse("tel:03-331-6005;phone-context=+64;a=%A1", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Test parsing RFC3966 with an ISDN subaddress.
	got, err = Parse("tel:03-331-6005;isub=12345;phone-context=+64", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("tel:+64-3-331-6005;isub=12345", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Test parsing RFC3966 with "tel:" missing.
	got, err = Parse("03-331-6005;phone-context=+64", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Testing international prefixes.
	// Should strip country calling code.
	got, err = Parse("0064 3 331 6005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Try again, but this time we have an international number with Region Code US. It should
	// recognise the country calling code and parse accordingly.
	got, err = Parse("01164 3 331 6005", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("+64 3 331 6005", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// We should ignore the leading plus here, since it is not followed by a valid country code but
	// instead is followed by the IDD for the US.
	got, err = Parse("+01164 3 331 6005", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("+0064 3 331 6005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("+ 00 64 3 331 6005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))

	got, err = Parse("tel:253-0000;phone-context=www.google.com", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usLocalNumber(), got))
	got, err = Parse("tel:253-0000;isub=12345;phone-context=www.google.com", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usLocalNumber(), got))
	got, err = Parse("tel:2530000;isub=12345;phone-context=1234.com", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usLocalNumber(), got))

	nzNumberInline := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(64123456)}
	got, err = Parse("64(0)64123456", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumberInline, got))
	// Check that using a "/" is fine in a phone number.
	got, err = Parse("301/23456", regionCode.DE)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(deNumber(), got))

	// Check it doesn't use the '1' as a country calling code when parsing if the phone number was
	// already possible.
	usNumberInline := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(1234567890)}
	got, err = Parse("123-456-7890", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumberInline, got))

	// Test star numbers. Although this is not strictly valid, we would like to make sure we can
	// parse the output we produce when formatting the number.
	got, err = Parse("+81 *2345", regionCode.JP)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(jpStarNumber(), got))

	shortNumber := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(12)}
	got, err = Parse("12", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(shortNumber, got))

	// Test for short-code with leading zero for a country which has 0 as national prefix. Ensure
	// it's not interpreted as national prefix if the remaining number length is local-only in
	// terms of length. Example: In GB, length 6-7 are only possible local-only.
	shortNumber = &PhoneNumber{
		CountryCode:        proto.Int32(44),
		NationalNumber:     proto.Uint64(123456),
		ItalianLeadingZero: proto.Bool(true),
	}
	got, err = Parse("0123456", regionCode.GB)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(shortNumber, got))
}

func TestParseNumberWithAlphaCharacters(t *testing.T) {
	useTestMetadata(t)

	// Test case with alpha characters.
	tollfreeNumber := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(800332005)}
	got, err := Parse("0800 DDA 005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(tollfreeNumber, got))
	premiumNumber := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(9003326005)}
	got, err = Parse("0900 DDA 6005", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(premiumNumber, got))
	// Not enough alpha characters for them to be considered intentional, so they are stripped.
	got, err = Parse("0900 332 6005a", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(premiumNumber, got))
	got, err = Parse("0900 332 600a5", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(premiumNumber, got))
	got, err = Parse("0900 332 600A5", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(premiumNumber, got))
	got, err = Parse("0900 a332 600A5", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(premiumNumber, got))
}

func TestParseMaliciousInput(t *testing.T) {
	useTestMetadata(t)

	// Lots of leading + signs before the possible number.
	var maliciousNumber strings.Builder
	for i := 0; i < 6000; i++ {
		maliciousNumber.WriteByte('+')
	}
	maliciousNumber.WriteString("12222-33-244 extensioB 343+")
	_, err := Parse(maliciousNumber.String(), regionCode.US)
	assert.ErrorIs(t, err, ErrNumTooLong)

	var maliciousNumberWithAlmostExt strings.Builder
	for i := 0; i < 350; i++ {
		maliciousNumberWithAlmostExt.WriteString("200")
	}
	maliciousNumberWithAlmostExt.WriteString(" extensiOB 345")
	_, err = Parse(maliciousNumberWithAlmostExt.String(), regionCode.US)
	assert.ErrorIs(t, err, ErrNumTooLong)
}

func TestParseWithInternationalPrefixes(t *testing.T) {
	useTestMetadata(t)

	got, err := Parse("+1 (650) 253-0000", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	got, err = Parse("011 800 1234 5678", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(internationalTollFree(), got))
	got, err = Parse("1-650-253-0000", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// Calling the US number from Singapore by using different service providers
	// 1st test: calling using SingTel IDD service (IDD is 001)
	got, err = Parse("0011-650-253-0000", regionCode.SG)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// 2nd test: calling using StarHub IDD service (IDD is 008)
	got, err = Parse("0081-650-253-0000", regionCode.SG)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// 3rd test: calling using SingTel V019 service (IDD is 019)
	got, err = Parse("0191-650-253-0000", regionCode.SG)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// Calling the US number from Poland
	got, err = Parse("0~01-650-253-0000", regionCode.PL)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// Using "++" at the start.
	got, err = Parse("++1 (650) 253-0000", regionCode.PL)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
}

func TestParseNonAscii(t *testing.T) {
	useTestMetadata(t)

	// Using a full-width plus sign.
	got, err := Parse("＋1 (650) 253-0000", regionCode.SG)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// Using a soft hyphen U+00AD.
	got, err = Parse("1 (650) 253­-0000", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// The whole number, including punctuation, is here represented in full-width form.
	got, err = Parse("＋１　（６５０）"+
		"　２５３－００００",
		regionCode.SG)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))
	// Using U+30FC dash instead.
	got, err = Parse("＋１　（６５０）"+
		"　２５３ー００００",
		regionCode.SG)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber(), got))

	// Using a very strange decimal digit range (Mongolian digits).
	// DIVERGENCE(upstream): our normalizeDigits only maps ASCII and Arabic-Indic
	// numerals to ASCII digits; it leaves other unicode.IsDigit runes (e.g.
	// Mongolian U+1810-U+1819) as the raw multi-byte rune rather than converting
	// to '0'-'9' as Java's Character.digit does. The number therefore fails to
	// parse (reported as too long) instead of yielding US_NUMBER. — TODO reconcile
	_, err = Parse("᠑ ᠖᠕᠐ "+
		"᠒᠕᠓ ᠐᠐᠐᠐",
		regionCode.US)
	assert.ErrorIs(t, err, ErrNumTooLong)
}

func TestParseWithLeadingZero(t *testing.T) {
	useTestMetadata(t)

	got, err := Parse("+39 02-36618 300", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(itNumber(), got))
	got, err = Parse("02-36618 300", regionCode.IT)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(itNumber(), got))

	got, err = Parse("345 678 901", regionCode.IT)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(itMobile(), got))
}

func TestParseNationalNumberArgentina(t *testing.T) {
	useTestMetadata(t)

	// Test parsing mobile numbers of Argentina.
	arNumberInline := &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(93435551212)}
	got, err := Parse("+54 9 343 555 1212", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))
	got, err = Parse("0343 15 555 1212", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))

	arNumberInline = &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(93715654320)}
	got, err = Parse("+54 9 3715 65 4320", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))
	got, err = Parse("03715 15 65 4320", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))
	got, err = Parse("911 876 54321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arMobile(), got))

	// Test parsing fixed-line numbers of Argentina.
	got, err = Parse("+54 11 8765 4321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumber(), got))
	got, err = Parse("011 8765 4321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumber(), got))

	arNumberInline = &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(3715654321)}
	got, err = Parse("+54 3715 65 4321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))
	got, err = Parse("03715 65 4321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))

	arNumberInline = &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(2312340000)}
	got, err = Parse("+54 23 1234 0000", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))
	got, err = Parse("023 1234 0000", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumberInline, got))
}

func TestParseWithXInNumber(t *testing.T) {
	useTestMetadata(t)

	// Test that having an 'x' in the phone number at the start is ok and that it just gets removed.
	got, err := Parse("01187654321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumber(), got))
	got, err = Parse("(0) 1187654321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumber(), got))
	got, err = Parse("0 1187654321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumber(), got))
	got, err = Parse("(0xx) 1187654321", regionCode.AR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arNumber(), got))
	arFromUs := &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(81429712)}
	// This test is intentionally constructed such that the number of digit after xx is larger than
	// 7, so that the number won't be mistakenly treated as an extension, as we allow extensions up
	// to 7 digits. This assumption is okay for now as all the countries where a carrier selection
	// code is written in the form of xx have a national significant number of length larger than 7.
	got, err = Parse("011xx5481429712", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(arFromUs, got))
}

func TestParseNumbersMexico(t *testing.T) {
	useTestMetadata(t)

	// Test parsing fixed-line numbers of Mexico.
	mxNumber := &PhoneNumber{CountryCode: proto.Int32(52), NationalNumber: proto.Uint64(4499780001)}
	got, err := Parse("+52 (449)978-0001", regionCode.MX)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(mxNumber, got))
	got, err = Parse("01 (449)978-0001", regionCode.MX)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(mxNumber, got))
	got, err = Parse("(449)978-0001", regionCode.MX)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(mxNumber, got))

	// Test parsing mobile numbers of Mexico.
	mxNumber = &PhoneNumber{CountryCode: proto.Int32(52), NationalNumber: proto.Uint64(13312345678)}
	got, err = Parse("+52 1 33 1234-5678", regionCode.MX)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(mxNumber, got))
	got, err = Parse("044 (33) 1234-5678", regionCode.MX)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(mxNumber, got))
	got, err = Parse("045 33 1234-5678", regionCode.MX)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(mxNumber, got))
}

func TestFailedParseOnInvalidNumbers(t *testing.T) {
	useTestMetadata(t)

	tests := []struct {
		name   string
		input  string
		region string
		err    error
	}{
		{"sentence", "This is not a phone number", regionCode.NZ, ErrNotANumber},
		{"oneAndSentence", "1 Still not a number", regionCode.NZ, ErrNotANumber},
		{"oneMicrosoft", "1 MICROSOFT", regionCode.NZ, ErrNotANumber},
		{"twelveMicrosoft", "12 MICROSOFT", regionCode.NZ, ErrNotANumber},
		{"tooLong", "01495 72553301873 810104", regionCode.GB, ErrNumTooLong},
		{"plusMinus", "+---", regionCode.DE, ErrNotANumber},
		{"plusStar", "+***", regionCode.DE, ErrNotANumber},
		{"plusStarNumber", "+*******91", regionCode.DE, ErrNotANumber},
		{"tooShort", "+49 0", regionCode.DE, ErrTooShortNSN},
		{"invalidCountryCode", "+210 3456 56789", regionCode.NZ, ErrInvalidCountryCode},
		{"plusAndIddAndInvalidCountryCode", "+ 00 210 3 331 6005", regionCode.NZ, ErrInvalidCountryCode},
		{"unknownRegion", "123 456 7890", regionCode.ZZ, ErrInvalidCountryCode},
		{"deprecatedRegion", "123 456 7890", "CS", ErrInvalidCountryCode},
		{"nullRegion", "123 456 7890", "", ErrInvalidCountryCode},
		{"onlyRegionCodeDashes", "0044------", regionCode.GB, ErrTooShortAfterIDD},
		{"onlyRegionCode", "0044", regionCode.GB, ErrTooShortAfterIDD},
		{"onlyIdd", "011", regionCode.US, ErrTooShortAfterIDD},
		{"onlyIddThen9", "0119", regionCode.US, ErrTooShortAfterIDD},
		{"emptyZZ", "", regionCode.ZZ, ErrNotANumber},
		{"empty string US", "", regionCode.US, ErrNotANumber},
		{"domainRfcPhoneContextZZ", "tel:555-1234;phone-context=www.google.com", regionCode.ZZ, ErrInvalidCountryCode},
		{"rfcPhoneContextNoPlus", "tel:555-1234;phone-context=1-331", regionCode.ZZ, ErrNotANumber},
		{"rfcPhoneContextEmpty", ";phone-context=", regionCode.ZZ, ErrNotANumber},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.input, tc.region)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestParseNumbersWithPlusWithNoRegion(t *testing.T) {
	useTestMetadata(t)

	// regionCode.ZZ is allowed only if the number starts with a '+' - then the country calling code
	// can be calculated.
	got, err := Parse("+64 3 331 6005", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Test with full-width plus.
	got, err = Parse("＋64 3 331 6005", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	// Test with normal plus but leading characters that need to be stripped.
	got, err = Parse("Tel: +64 3 331 6005", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("+64 3 331 6005", "")
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("+800 1234 5678", "")
	assert.NoError(t, err)
	assert.True(t, proto.Equal(internationalTollFree(), got))
	got, err = Parse("+979 123 456 789", "")
	assert.NoError(t, err)
	assert.True(t, proto.Equal(universalPremiumRate(), got))

	// Test parsing RFC3966 format with a phone context.
	got, err = Parse("tel:03-331-6005;phone-context=+64", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("  tel:03-331-6005;phone-context=+64", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))
	got, err = Parse("tel:03-331-6005;isub=12345;phone-context=+64", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))

	nzNumberWithRawInput := nzNumber()
	nzNumberWithRawInput.RawInput = proto.String("+64 3 331 6005")
	nzNumberWithRawInput.CountryCodeSource = PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN.Enum()
	got, err = ParseAndKeepRawInput("+64 3 331 6005", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumberWithRawInput, got))
	// Null is also allowed for the region code in these cases.
	got, err = ParseAndKeepRawInput("+64 3 331 6005", "")
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumberWithRawInput, got))
}

func TestParseNumberTooShortIfNationalPrefixStripped(t *testing.T) {
	useTestMetadata(t)

	// Test that a number whose first digits happen to coincide with the national prefix does not
	// get them stripped if doing so would result in a number too short to be a possible (regular
	// length) phone number for that region.
	byNumber := &PhoneNumber{CountryCode: proto.Int32(375), NationalNumber: proto.Uint64(8123)}
	got, err := Parse("8123", regionCode.BY)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(byNumber, got))
	byNumber.NationalNumber = proto.Uint64(81234)
	got, err = Parse("81234", regionCode.BY)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(byNumber, got))

	// The prefix doesn't get stripped, since the input is a viable 6-digit number, whereas the
	// result of stripping is only 5 digits.
	byNumber.NationalNumber = proto.Uint64(812345)
	got, err = Parse("812345", regionCode.BY)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(byNumber, got))

	// The prefix gets stripped, since only 6-digit numbers are possible.
	byNumber.NationalNumber = proto.Uint64(123456)
	got, err = Parse("8123456", regionCode.BY)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(byNumber, got))
}

func TestParseExtensions(t *testing.T) {
	useTestMetadata(t)

	nzNumber := pn(64, 33316005)
	nzNumber.Extension = proto.String("3456")
	for _, in := range []string{
		"03 331 6005 ext 3456",
		"03-3316005x3456",
		"03-3316005 int.3456",
		"03 3316005 #3456",
	} {
		got, err := Parse(in, regionCode.NZ)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(nzNumber, got), "input %q", in)
	}

	// Test the following do not extract extensions:
	for _, tc := range []struct{ in, region string }{
		{"1800 six-flags", regionCode.US},
		{"1800 SIX FLAGS", regionCode.US},
		{"0~0 1800 7493 5247", regionCode.PL},
		{"(1800) 7493.5247", regionCode.US},
	} {
		got, err := Parse(tc.in, tc.region)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(alphaNumericNumber(), got), "input %q", tc.in)
	}

	// Check that the last instance of an extension token is matched.
	extnNumber := alphaNumericNumber()
	extnNumber.Extension = proto.String("1234")
	got, err := Parse("0~0 1800 7493 5247 ~1234", regionCode.PL)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(extnNumber, got))

	// Verifying bug-fix where the last digit of a number was previously omitted if it was a 0 when
	// extracting the extension. Also verifying a few different cases of extensions.
	ukNumber := pn(44, 2034567890)
	ukNumber.Extension = proto.String("456")
	for _, tc := range []struct{ in, region string }{
		{"+44 2034567890x456", regionCode.NZ},
		{"+44 2034567890x456", regionCode.GB},
		{"+44 2034567890 x456", regionCode.GB},
		{"+44 2034567890 X456", regionCode.GB},
		{"+44 2034567890 X 456", regionCode.GB},
		{"+44 2034567890 X  456", regionCode.GB},
		{"+44 2034567890  X 456", regionCode.GB},
		{"+44 2034567890 x 456  ", regionCode.GB},
		{"+44-2034567890;ext=456", regionCode.GB},
		{"tel:2034567890;ext=456;phone-context=+44", regionCode.ZZ},
		// Full-width extension, "extn" only.
		{"+442034567890ｅｘｔｎ456", regionCode.GB},
		// "xtn" only.
		{"+442034567890ｘｔｎ456", regionCode.GB},
		// "xt" only.
		{"+442034567890ｘｔ456", regionCode.GB},
	} {
		got, err := Parse(tc.in, tc.region)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(ukNumber, got), "input %q", tc.in)
	}

	usWithExtension := pn(1, 8009013355)
	usWithExtension.Extension = proto.String("7246433")
	for _, in := range []string{
		"(800) 901-3355 x 7246433",
		"(800) 901-3355 , ext 7246433",
		"(800) 901-3355 ; 7246433",
		// To test an extension character without surrounding spaces.
		"(800) 901-3355;7246433",
		"(800) 901-3355 ,extension 7246433",
		"(800) 901-3355 ,extensión 7246433",
		// Repeat with the small letter o with acute accent created by combining characters.
		"(800) 901-3355 ,extensión 7246433",
		"(800) 901-3355 , 7246433",
		"(800) 901-3355 ext: 7246433",
	} {
		got, err := Parse(in, regionCode.US)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(usWithExtension, got), "input %q", in)
	}

	// Testing Russian extension доб with variants found online.
	ruWithExtension := pn(7, 4232022511)
	ruWithExtension.Extension = proto.String("100")
	for _, in := range []string{
		"8 (423) 202-25-11, доб. 100",
		"8 (423) 202-25-11 доб. 100",
		"8 (423) 202-25-11, доб 100",
		"8 (423) 202-25-11 доб 100",
		"8 (423) 202-25-11доб100",
		// In upper case
		"8 (423) 202-25-11, ДОБ. 100",
	} {
		got, err := Parse(in, regionCode.RU)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(ruWithExtension, got), "input %q", in)
	}

	// Test that if a number has two extensions specified, we ignore the second.
	usWithTwoExtensionsNumber := pn(1, 2121231234)
	usWithTwoExtensionsNumber.Extension = proto.String("508")
	for _, in := range []string{
		"(212)123-1234 x508/x1234",
		"(212)123-1234 x508/ x1234",
		"(212)123-1234 x508\\x1234",
	} {
		got, err := Parse(in, regionCode.US)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(usWithTwoExtensionsNumber, got), "input %q", in)
	}

	// Test parsing numbers in the form (645) 123-1234-910# works, where the last 3 digits before
	// the # are an extension.
	usWithExtension2 := pn(1, 6451231234)
	usWithExtension2.Extension = proto.String("910")
	for _, in := range []string{
		"+1 (645) 123 1234-910#",
		// Retry with the same number in a slightly different format.
		"+1 (645) 123 1234 ext. 910#",
	} {
		got, err := Parse(in, regionCode.US)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(usWithExtension2, got), "input %q", in)
	}
}

func TestParseHandlesLongExtensionsWithExplicitLabels(t *testing.T) {
	useTestMetadata(t)

	// Test lower and upper limits of extension lengths for each type of label.
	nzNumber := pn(64, 33316005)

	// Firstly, when in RFC format: extLimitAfterExplicitLabel
	nzNumber.Extension = proto.String("0")
	got, err := Parse("tel:+6433316005;ext=0", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber, got))

	nzNumber.Extension = proto.String("01234567890123456789")
	got, err = Parse("tel:+6433316005;ext=01234567890123456789", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber, got))

	// Extension too long.
	_, err = Parse("tel:+6433316005;ext=012345678901234567890", regionCode.NZ)
	assert.ErrorIs(t, err, ErrNotANumber)

	// Explicit extension label: extLimitAfterExplicitLabel
	nzNumber.Extension = proto.String("1")
	got, err = Parse("03 3316005ext:1", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber, got))

	nzNumber.Extension = proto.String("12345678901234567890")
	for _, in := range []string{
		"03 3316005 xtn:12345678901234567890",
		"03 3316005 extension\t12345678901234567890",
		"03 3316005 xtensio:12345678901234567890",
		"03 3316005 xtensión, 12345678901234567890#",
		"03 3316005extension.12345678901234567890",
		"03 3316005 доб:12345678901234567890",
	} {
		got, err := Parse(in, regionCode.NZ)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(nzNumber, got), "input %q", in)
	}

	// Extension too long.
	_, err = Parse("03 3316005 extension 123456789012345678901", regionCode.NZ)
	assert.ErrorIs(t, err, ErrNumTooLong)
}

func TestParseHandlesLongExtensionsWithAutoDiallingLabels(t *testing.T) {
	useTestMetadata(t)

	// Secondly, cases of auto-dialling and other standard extension labels,
	// extLimitAfterLikelyLabel
	usNumberUserInput := pn(1, 2679000000)
	usNumberUserInput.Extension = proto.String("123456789012345")
	for _, in := range []string{
		"+12679000000,,123456789012345#",
		"+12679000000;123456789012345#",
	} {
		got, err := Parse(in, regionCode.US)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(usNumberUserInput, got), "input %q", in)
	}

	ukNumberUserInput := pn(44, 2034000000)
	ukNumberUserInput.Extension = proto.String("123456789")
	got, err := Parse("+442034000000,,123456789#", regionCode.GB)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(ukNumberUserInput, got))

	// Extension too long.
	_, err = Parse("+12679000000,,1234567890123456#", regionCode.US)
	assert.ErrorIs(t, err, ErrNotANumber)
}

func TestParseHandlesShortExtensionsWithAmbiguousChar(t *testing.T) {
	useTestMetadata(t)

	nzNumber := pn(64, 33316005)

	// Thirdly, for single and non-standard cases: extLimitAfterAmbiguousChar
	nzNumber.Extension = proto.String("123456789")
	for _, in := range []string{
		"03 3316005 x 123456789",
		"03 3316005 x. 123456789",
		"03 3316005 #123456789#",
		"03 3316005 ~ 123456789",
	} {
		got, err := Parse(in, regionCode.NZ)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(nzNumber, got), "input %q", in)
	}

	// Extension too long.
	_, err := Parse("03 3316005 ~ 1234567890", regionCode.NZ)
	assert.ErrorIs(t, err, ErrNumTooLong)
}

func TestParseHandlesShortExtensionsWhenNotSureOfLabel(t *testing.T) {
	useTestMetadata(t)

	// Lastly, when no explicit extension label present, but denoted by tailing #:
	// extLimitWhenNotSure
	usNumber := pn(1, 1234567890)
	usNumber.Extension = proto.String("666666")
	got, err := Parse("+1123-456-7890 666666#", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber, got))

	usNumber.Extension = proto.String("6")
	got, err = Parse("+11234567890-6#", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usNumber, got))

	// Extension too long.
	_, err = Parse("+1123-456-7890 7777777#", regionCode.US)
	assert.ErrorIs(t, err, ErrNotANumber)
}

func TestParseAndKeepRaw(t *testing.T) {
	useTestMetadata(t)

	alpha := alphaNumericNumber()
	alpha.RawInput = proto.String("800 six-flags")
	alpha.CountryCodeSource = PhoneNumber_FROM_DEFAULT_COUNTRY.Enum()
	got, err := ParseAndKeepRawInput("800 six-flags", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(alpha, got))

	shorterAlphaNumber := pn(1, 8007493524)
	shorterAlphaNumber.RawInput = proto.String("1800 six-flag")
	shorterAlphaNumber.CountryCodeSource = PhoneNumber_FROM_NUMBER_WITHOUT_PLUS_SIGN.Enum()
	got, err = ParseAndKeepRawInput("1800 six-flag", regionCode.US)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(shorterAlphaNumber, got))

	shorterAlphaNumber.RawInput = proto.String("+1800 six-flag")
	shorterAlphaNumber.CountryCodeSource = PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN.Enum()
	got, err = ParseAndKeepRawInput("+1800 six-flag", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(shorterAlphaNumber, got))

	shorterAlphaNumber.RawInput = proto.String("001800 six-flag")
	shorterAlphaNumber.CountryCodeSource = PhoneNumber_FROM_NUMBER_WITH_IDD.Enum()
	got, err = ParseAndKeepRawInput("001800 six-flag", regionCode.NZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(shorterAlphaNumber, got))

	// Invalid region code supplied.
	_, err = ParseAndKeepRawInput("123 456 7890", "CS")
	assert.ErrorIs(t, err, ErrInvalidCountryCode)

	koreanNumber := pn(82, 22123456)
	koreanNumber.RawInput = proto.String("08122123456")
	koreanNumber.CountryCodeSource = PhoneNumber_FROM_DEFAULT_COUNTRY.Enum()
	koreanNumber.PreferredDomesticCarrierCode = proto.String("81")
	got, err = ParseAndKeepRawInput("08122123456", regionCode.KR)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(koreanNumber, got))
}

func TestParseItalianLeadingZeros(t *testing.T) {
	useTestMetadata(t)

	// Test the number "011".
	oneZero := &PhoneNumber{
		CountryCode:        proto.Int32(61),
		NationalNumber:     proto.Uint64(11),
		ItalianLeadingZero: proto.Bool(true),
	}
	got, err := Parse("011", regionCode.AU)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(oneZero, got))

	// Test the number "001".
	twoZeros := &PhoneNumber{
		CountryCode:          proto.Int32(61),
		NationalNumber:       proto.Uint64(1),
		ItalianLeadingZero:   proto.Bool(true),
		NumberOfLeadingZeros: proto.Int32(2),
	}
	got, err = Parse("001", regionCode.AU)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(twoZeros, got))

	// Test the number "000". This number has 2 leading zeros.
	stillTwoZeros := &PhoneNumber{
		CountryCode:          proto.Int32(61),
		NationalNumber:       proto.Uint64(0),
		ItalianLeadingZero:   proto.Bool(true),
		NumberOfLeadingZeros: proto.Int32(2),
	}
	got, err = Parse("000", regionCode.AU)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(stillTwoZeros, got))

	// Test the number "0000". This number has 3 leading zeros.
	threeZeros := &PhoneNumber{
		CountryCode:          proto.Int32(61),
		NationalNumber:       proto.Uint64(0),
		ItalianLeadingZero:   proto.Bool(true),
		NumberOfLeadingZeros: proto.Int32(3),
	}
	got, err = Parse("0000", regionCode.AU)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(threeZeros, got))
}

func TestParseWithPhoneContext(t *testing.T) {
	useTestMetadata(t)

	// context    = ";phone-context=" descriptor
	// descriptor = domainname / global-number-digits

	// Valid global-phone-digits
	got, err := Parse("tel:033316005;phone-context=+64", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))

	got, err = Parse("tel:033316005;phone-context=+64;{this isn't part of phone-context anymore!}", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzNumber(), got))

	nzFromPhoneContext := pn(64, 3033316005)
	got, err = Parse("tel:033316005;phone-context=+64-3", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(nzFromPhoneContext, got))

	brFromPhoneContext := pn(55, 5033316005)
	got, err = Parse("tel:033316005;phone-context=+(555)", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(brFromPhoneContext, got))

	usFromPhoneContext := pn(1, 23033316005)
	got, err = Parse("tel:033316005;phone-context=+-1-2.3()", regionCode.ZZ)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(usFromPhoneContext, got))

	// Valid domainname
	for _, in := range []string{
		"tel:033316005;phone-context=abc.nz",
		"tel:033316005;phone-context=www.PHONE-numb3r.com",
		"tel:033316005;phone-context=a",
		"tel:033316005;phone-context=3phone.J.",
		"tel:033316005;phone-context=a--z",
	} {
		got, err := Parse(in, regionCode.NZ)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(nzNumber(), got), "input %q", in)
	}

	// Invalid descriptor
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=+")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=64")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=++64")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=+abc")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=.")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=3phone")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=a-.nz")
	assertThrowsForInvalidPhoneContext(t, "tel:033316005;phone-context=a{b}c")
}

func TestCountryWithNoNumberDesc(t *testing.T) {
	useTestMetadata(t)

	// Andorra is a country where we don't have PhoneNumberDesc info in the metadata.
	adNumber := pn(376, 12345)
	assert.Equal(t, "+376 12345", Format(adNumber, INTERNATIONAL))
	assert.Equal(t, "+37612345", Format(adNumber, E164))
	assert.Equal(t, "12345", Format(adNumber, NATIONAL))
	assert.Equal(t, UNKNOWN, GetNumberType(adNumber))
	assert.False(t, IsValidNumber(adNumber))

	// Test dialing a US number from within Andorra.
	assert.Equal(t, "00 1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.AD))
}

func TestUnknownCountryCallingCode(t *testing.T) {
	useTestMetadata(t)

	assert.False(t, IsValidNumber(unknownCountryCodeNoRawInput()))
	// It's not very well defined as to what the E164 representation for a number with an invalid
	// country calling code is, but just prefixing the country code and national number is about
	// the best we can do.
	assert.Equal(t, "+212345", Format(unknownCountryCodeNoRawInput(), E164))
}

// testIsNumberMatchMatches (PhoneNumberUtilTest.java:2992-3036)
func TestIsNumberMatchMatches(t *testing.T) {
	useTestMetadata(t)
	// Test simple matches where formatting is different, or leading zeros, or country calling code
	// has been specified.
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+64 3 331 6005", "+64 03 331 6005"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+800 1234 5678", "+80012345678"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+64 03 331-6005", "+64 03331 6005"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+643 331-6005", "+64033316005"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+643 331-6005", "+6433316005"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+64 3 331-6005", "+6433316005"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+64 3 331-6005", "tel:+64-3-331-6005;isub=123"))
	// Test alpha numbers.
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+1800 siX-Flags", "+1 800 7493 5247"))
	// Test numbers with extensions.
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+64 3 331-6005 extn 1234", "+6433316005#1234"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+64 3 331-6005 ext. 1234", "+6433316005;1234"))
	assert.Equal(t, EXACT_MATCH, IsNumberMatch("+7 423 202-25-11 ext 100", "+7 4232022511 доб. 100"))
	// Test proto buffers.
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithOneNumber(nzNumber(), "+6403 331 6005"))

	nz := nzNumber()
	nz.Extension = proto.String("3456")
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithOneNumber(nz, "+643 331 6005 ext 3456"))
	// Check empty extensions are ignored.
	nz.Extension = proto.String("")
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithOneNumber(nz, "+6403 331 6005"))
	// Check variant with two proto buffers.
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithNumbers(nz, nzNumber()))
}

// testIsNumberMatchShortMatchIfDiffNumLeadingZeros (PhoneNumberUtilTest.java:3038-3053)
func TestIsNumberMatchShortMatchIfDiffNumLeadingZeros(t *testing.T) {
	useTestMetadata(t)
	nzNumberOne := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(33316005), ItalianLeadingZero: proto.Bool(true)}
	nzNumberTwo := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(33316005), ItalianLeadingZero: proto.Bool(true), NumberOfLeadingZeros: proto.Int32(2)}
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatchWithNumbers(nzNumberOne, nzNumberTwo))

	// Since one doesn't have the "italian_leading_zero" set to true, we ignore the number of
	// leading zeros present (1 is in any case the default value).
	nzNumberOne.ItalianLeadingZero = proto.Bool(false)
	nzNumberOne.NumberOfLeadingZeros = proto.Int32(1)
	nzNumberTwo.ItalianLeadingZero = proto.Bool(true)
	nzNumberTwo.NumberOfLeadingZeros = proto.Int32(1)
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatchWithNumbers(nzNumberOne, nzNumberTwo))
}

// testIsNumberMatchAcceptsProtoDefaultsAsMatch (PhoneNumberUtilTest.java:3055-3065)
func TestIsNumberMatchAcceptsProtoDefaultsAsMatch(t *testing.T) {
	useTestMetadata(t)
	nzNumberOne := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(33316005), ItalianLeadingZero: proto.Bool(true)}
	// The default for number_of_leading_zeros is 1, so it shouldn't normally be set, however if it
	// is it should be considered equivalent.
	nzNumberTwo := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(33316005), ItalianLeadingZero: proto.Bool(true), NumberOfLeadingZeros: proto.Int32(1)}
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithNumbers(nzNumberOne, nzNumberTwo))
}

// testIsNumberMatchMatchesDiffLeadingZerosIfItalianLeadingZeroFalse (PhoneNumberUtilTest.java:3067-3082)
func TestIsNumberMatchMatchesDiffLeadingZerosIfItalianLeadingZeroFalse(t *testing.T) {
	useTestMetadata(t)
	nzNumberOne := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(33316005)}
	// The default for number_of_leading_zeros is 1, so it shouldn't normally be set, however if it
	// is it should be considered equivalent.
	nzNumberTwo := &PhoneNumber{CountryCode: proto.Int32(64), NationalNumber: proto.Uint64(33316005), NumberOfLeadingZeros: proto.Int32(1)}
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithNumbers(nzNumberOne, nzNumberTwo))

	// Even if it is set to ten, it is still equivalent because in both cases
	// italian_leading_zero is not true.
	nzNumberTwo.NumberOfLeadingZeros = proto.Int32(10)
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithNumbers(nzNumberOne, nzNumberTwo))
}

// testIsNumberMatchIgnoresSomeFields (PhoneNumberUtilTest.java:3084-3096)
func TestIsNumberMatchIgnoresSomeFields(t *testing.T) {
	useTestMetadata(t)
	// Check raw_input, country_code_source and preferred_domestic_carrier_code are ignored.
	brNumberOne := &PhoneNumber{CountryCode: proto.Int32(55), NationalNumber: proto.Uint64(3121286979)}
	brNumberOne.CountryCodeSource = PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN.Enum()
	brNumberOne.PreferredDomesticCarrierCode = proto.String("12")
	brNumberOne.RawInput = proto.String("012 3121286979")
	brNumberTwo := &PhoneNumber{CountryCode: proto.Int32(55), NationalNumber: proto.Uint64(3121286979)}
	brNumberTwo.CountryCodeSource = PhoneNumber_FROM_DEFAULT_COUNTRY.Enum()
	brNumberTwo.PreferredDomesticCarrierCode = proto.String("14")
	brNumberTwo.RawInput = proto.String("143121286979")
	assert.Equal(t, EXACT_MATCH, IsNumberMatchWithNumbers(brNumberOne, brNumberTwo))
}

// testIsNumberMatchNonMatches (PhoneNumberUtilTest.java:3098-3129)
func TestIsNumberMatchNonMatches(t *testing.T) {
	useTestMetadata(t)
	// Non-matches.
	assert.Equal(t, NO_MATCH, IsNumberMatch("03 331 6005", "03 331 6006"))
	assert.Equal(t, NO_MATCH, IsNumberMatch("+800 1234 5678", "+1 800 1234 5678"))
	// Different country calling code, partial number match.
	assert.Equal(t, NO_MATCH, IsNumberMatch("+64 3 331-6005", "+16433316005"))
	// Different country calling code, same number.
	assert.Equal(t, NO_MATCH, IsNumberMatch("+64 3 331-6005", "+6133316005"))
	// Extension different, all else the same.
	assert.Equal(t, NO_MATCH, IsNumberMatch("+64 3 331-6005 extn 1234", "0116433316005#1235"))
	assert.Equal(t, NO_MATCH, IsNumberMatch("+64 3 331-6005 extn 1234", "tel:+64-3-331-6005;ext=1235"))
	// NSN matches, but extension is different - not the same number.
	assert.Equal(t, NO_MATCH, IsNumberMatch("+64 3 331-6005 ext.1235", "3 331 6005#1234"))

	// Invalid numbers that can't be parsed.
	assert.Equal(t, NOT_A_NUMBER, IsNumberMatch("4", "3 331 6043"))
	assert.Equal(t, NOT_A_NUMBER, IsNumberMatch("+43", "+64 3 331 6005"))
	assert.Equal(t, NOT_A_NUMBER, IsNumberMatch("+43", "64 3 331 6005"))
	assert.Equal(t, NOT_A_NUMBER, IsNumberMatch("Dog", "64 3 331 6005"))
}

// testIsNumberMatchNsnMatches (PhoneNumberUtilTest.java:3131-3166)
func TestIsNumberMatchNsnMatches(t *testing.T) {
	useTestMetadata(t)
	// NSN matches.
	assert.Equal(t, NSN_MATCH, IsNumberMatch("+64 3 331-6005", "03 331 6005"))
	assert.Equal(t, NSN_MATCH, IsNumberMatch("+64 3 331-6005", "tel:03-331-6005;isub=1234;phone-context=abc.nz"))
	assert.Equal(t, NSN_MATCH, IsNumberMatchWithOneNumber(nzNumber(), "03 331 6005"))
	// Here the second number possibly starts with the country calling code for New Zealand,
	// although we are unsure.
	unchangedNzNumber := nzNumber()
	assert.Equal(t, NSN_MATCH, IsNumberMatchWithOneNumber(unchangedNzNumber, "(64-3) 331 6005"))
	// Check the phone number proto was not edited during the method call.
	assert.True(t, proto.Equal(nzNumber(), unchangedNzNumber))

	// Here, the 1 might be a national prefix, if we compare it to the US number, so the resultant
	// match is an NSN match.
	assert.Equal(t, NSN_MATCH, IsNumberMatchWithOneNumber(usNumber(), "1-650-253-0000"))
	assert.Equal(t, NSN_MATCH, IsNumberMatchWithOneNumber(usNumber(), "6502530000"))
	assert.Equal(t, NSN_MATCH, IsNumberMatch("+1 650-253 0000", "1 650 253 0000"))
	assert.Equal(t, NSN_MATCH, IsNumberMatch("1 650-253 0000", "1 650 253 0000"))
	assert.Equal(t, NSN_MATCH, IsNumberMatch("1 650-253 0000", "+1 650 253 0000"))
	// For this case, the match will be a short NSN match, because we cannot assume that the 1 might
	// be a national prefix, so don't remove it when parsing.
	randomNumber := &PhoneNumber{CountryCode: proto.Int32(41), NationalNumber: proto.Uint64(6502530000)}
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatchWithOneNumber(randomNumber, "1-650-253-0000"))
}

// testIsNumberMatchShortNsnMatches (PhoneNumberUtilTest.java:3168-3212)
func TestIsNumberMatchShortNsnMatches(t *testing.T) {
	useTestMetadata(t)
	// Short NSN matches with the country not specified for either one or both numbers.
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("+64 3 331-6005", "331 6005"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("+64 3 331-6005", "tel:331-6005;phone-context=abc.nz"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("+64 3 331-6005", "tel:331-6005;isub=1234;phone-context=abc.nz"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("+64 3 331-6005", "tel:331-6005;isub=1234;phone-context=abc.nz;a=%A1"))
	// We did not know that the "0" was a national prefix since neither number has a country code,
	// so this is considered a SHORT_NSN_MATCH.
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("3 331-6005", "03 331 6005"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("3 331-6005", "331 6005"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("3 331-6005", "tel:331-6005;phone-context=abc.nz"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("3 331-6005", "+64 331 6005"))
	// Short NSN match with the country specified.
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("03 331-6005", "331 6005"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("1 234 345 6789", "345 6789"))
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("+1 (234) 345 6789", "345 6789"))
	// NSN matches, country calling code omitted for one number, extension missing for one.
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatch("+64 3 331-6005", "3 331 6005#1234"))
	// One has Italian leading zero, one does not.
	italianNumberOne := &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(1234), ItalianLeadingZero: proto.Bool(true)}
	italianNumberTwo := &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(1234)}
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatchWithNumbers(italianNumberOne, italianNumberTwo))
	// One has an extension, the other has an extension of "".
	italianNumberOne.Extension = proto.String("1234")
	italianNumberOne.ItalianLeadingZero = nil // clearItalianLeadingZero()
	italianNumberTwo.Extension = proto.String("")
	assert.Equal(t, SHORT_NSN_MATCH, IsNumberMatchWithNumbers(italianNumberOne, italianNumberTwo))
}

func TestCanBeInternationallyDialled(t *testing.T) {
	useTestMetadata(t)
	// We have no-international-dialling rules for the US in our test metadata that
	// say that toll-free numbers cannot be dialled internationally.
	assert.False(t, CanBeInternationallyDialled(usTollFree()))

	// Normal US numbers can be internationally dialled.
	assert.True(t, CanBeInternationallyDialled(usNumber()))

	// Invalid number.
	assert.True(t, CanBeInternationallyDialled(usLocalNumber()))

	// We have no data for NZ - should return true.
	assert.True(t, CanBeInternationallyDialled(nzNumber()))
	assert.True(t, CanBeInternationallyDialled(internationalTollFree()))
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
