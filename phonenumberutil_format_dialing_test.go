package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java mobile-
// dialing, keep-alpha-chars, and in-original-format tests, run against the
// synthetic test metadata. Method names and assertions mirror the Java. Last
// reconciled against: v9.0.32

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// mustParseAndKeepRawInput parses and fails the test on error, mirroring the
// Java tests where parseAndKeepRawInput throws (and the test method declares
// "throws Exception"), so a parse regression surfaces as a clear failure rather
// than a nil dereference downstream.
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
