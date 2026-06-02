package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java, run
// against the synthetic test metadata (see testmetadata_test.go). Method names
// and assertions mirror the Java so this file can be kept in sync with upstream.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestMaybeStripNationalPrefixAndCarrierCode is the faithful port of upstream
// PhoneNumberUtilTest.testMaybeStripNationalPrefix. It builds its own metadata,
// so it doesn't need useTestMetadata.
func TestMaybeStripNationalPrefixAndCarrierCode(t *testing.T) {
	// Test basic national prefix stripping
	metadata := &PhoneMetadata{}
	metadata.NationalPrefixForParsing = proto.String("34")
	metadata.GeneralDesc = &PhoneNumberDesc{NationalNumberPattern: proto.String("\\d{4,8}")}

	number := NewStringBuilder([]byte("34356778"))
	assert.True(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "356778", number.String())

	// Retry - should not strip again
	assert.False(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "356778", number.String())

	// No national prefix
	metadata.NationalPrefixForParsing = proto.String("")
	number = NewStringBuilder([]byte("356778"))
	assert.False(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "356778", number.String())

	// If stripping doesn't match national rule, don't strip
	metadata.NationalPrefixForParsing = proto.String("3")
	number = NewStringBuilder([]byte("3123"))
	assert.False(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "3123", number.String())

	// Test extracting carrier code
	metadata.NationalPrefixForParsing = proto.String("0(81)?")
	number = NewStringBuilder([]byte("08122123456"))
	carrierCode := NewStringBuilder(nil)
	assert.True(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, carrierCode))
	assert.Equal(t, "81", carrierCode.String())
	assert.Equal(t, "22123456", number.String())

	// Test with transform rule
	metadata.NationalPrefixTransformRule = proto.String("5${1}5")
	metadata.NationalPrefixForParsing = proto.String("0(\\d{2})")
	number = NewStringBuilder([]byte("031123"))
	assert.True(t, maybeStripNationalPrefixAndCarrierCode(number, metadata, nil))
	assert.Equal(t, "5315123", number.String())
}

// testMaybeStripInternationalPrefix (PhoneNumberUtilTest.java:1902-1960)
func TestMaybeStripInternationalPrefix(t *testing.T) {
	internationalPrefix := "00[39]"
	numberToStrip := NewStringBuilder([]byte("0034567700-3898003"))
	// Note the dash is removed as part of the normalization.
	strippedNumber := "45677003898003"
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of its international prefix.")
	// Now the number no longer starts with an IDD prefix, so it should now report
	// FROM_DEFAULT_COUNTRY.
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))

	numberToStrip = NewStringBuilder([]byte("00945677003898003"))
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of its international prefix.")
	// Test it works when the international prefix is broken up by spaces.
	numberToStrip = NewStringBuilder([]byte("00 9 45677003898003"))
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of its international prefix.")
	// Now the number no longer starts with an IDD prefix, so it should now report
	// FROM_DEFAULT_COUNTRY.
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))

	// Test the + symbol is also recognised and stripped.
	numberToStrip = NewStringBuilder([]byte("+45677003898003"))
	strippedNumber = "45677003898003"
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied was not stripped of the plus symbol.")

	// If the number afterwards is a zero, we should not strip this - no country calling code begins
	// with 0.
	numberToStrip = NewStringBuilder([]byte("0090112-3123"))
	strippedNumber = "00901123123"
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
	assert.Equal(t, strippedNumber, numberToStrip.String(), "The number supplied had a 0 after the match so shouldn't be stripped.")
	// Here the 0 is separated by a space from the IDD.
	numberToStrip = NewStringBuilder([]byte("009 0-112-3123"))
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, maybeStripInternationalPrefixAndNormalize(numberToStrip, internationalPrefix))
}

// testMaybeExtractCountryCode (PhoneNumberUtilTest.java:1962-2090)
func TestMaybeExtractCountryCode(t *testing.T) {
	useTestMetadata(t)
	metadata := getMetadataForRegion(regionCode.US)

	// Note that for the US, the IDD is 011.
	number := &PhoneNumber{}
	numberToFill := NewStringBuilder(nil)
	cc, err := maybeExtractCountryCode("011112-3456789", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 1, cc, "Did not extract country calling code 1 correctly.")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_IDD, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")
	// Should strip and normalize national significant number.
	assert.Equal(t, "123456789", numberToFill.String(), "Did not strip off the country calling code correctly.")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
	cc, err = maybeExtractCountryCode("+6423456789", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 64, cc, "Did not extract country calling code 64 correctly.")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
	cc, err = maybeExtractCountryCode("+80012345678", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 800, cc, "Did not extract country calling code 800 correctly.")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
	cc, err = maybeExtractCountryCode("2345-6789", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 0, cc, "Should not have extracted a country calling code - no international prefix present.")
	assert.Equal(t, PhoneNumber_FROM_DEFAULT_COUNTRY, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
	_, err = maybeExtractCountryCode("0119991123456789", metadata, numberToFill, true, number)
	assert.ErrorIs(t, err, ErrInvalidCountryCode, "Should have thrown an exception, no valid country calling code present.")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
	cc, err = maybeExtractCountryCode("(1 610) 619 4466", metadata, numberToFill, true, number)
	assert.NoError(t, err)
	assert.Equal(t, 1, cc, "Should have extracted the country calling code of the region passed in")
	assert.Equal(t, PhoneNumber_FROM_NUMBER_WITHOUT_PLUS_SIGN, number.GetCountryCodeSource(), "Did not figure out CountryCodeSource correctly")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
	cc, err = maybeExtractCountryCode("(1 610) 619 4466", metadata, numberToFill, false, number)
	assert.NoError(t, err)
	assert.Equal(t, 1, cc, "Should have extracted the country calling code of the region passed in")
	assert.Nil(t, number.CountryCodeSource, "Should not contain CountryCodeSource.")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
	cc, err = maybeExtractCountryCode("(1 610) 619 446", metadata, numberToFill, false, number)
	assert.NoError(t, err)
	assert.Equal(t, 0, cc, "Should not have extracted a country calling code - invalid number after extraction of uncertain country calling code.")
	assert.Nil(t, number.CountryCodeSource, "Should not contain CountryCodeSource.")

	number = &PhoneNumber{}
	numberToFill = NewStringBuilder(nil)
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
