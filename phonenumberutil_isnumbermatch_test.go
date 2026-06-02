package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java
// testIsNumberMatch* methods, run against the synthetic test metadata (see
// testmetadata_test.go). Method names and assertions mirror the Java so this
// file can be kept in sync with upstream. Last reconciled against: v9.0.32
//
// Upstream's three isNumberMatch overloads map to:
//   isNumberMatch(CharSequence, CharSequence) -> IsNumberMatch
//   isNumberMatch(PhoneNumber, CharSequence)  -> IsNumberMatchWithOneNumber
//   isNumberMatch(PhoneNumber, PhoneNumber)   -> IsNumberMatchWithNumbers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

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
