package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java, run
// against the synthetic test metadata (see testmetadata_test.go). Covers the
// parsing-focused methods: invalid-number errors, extensions, keep-raw-input,
// and phone-context handling.
//
// Last reconciled against: v9.0.32

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

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

// assertThrowsForInvalidPhoneContext mirrors the private Java helper of the same
// name: parsing the number with an unknown region must fail with NOT_A_NUMBER.
func assertThrowsForInvalidPhoneContext(t *testing.T, numberToParse string) {
	_, err := Parse(numberToParse, regionCode.ZZ)
	assert.ErrorIs(t, err, ErrNotANumber, "input %q", numberToParse)
}
