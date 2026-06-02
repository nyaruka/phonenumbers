package phonenumbers

// Faithful port of upstream libphonenumber's AsYouTypeFormatterTest.java. As
// upstream notes, these tests use the synthetic test metadata, not the normal
// metadata file, so they are illustrative of functionality rather than a
// regression test of real-world data. Each test starts with useTestMetadata(t)
// (the analogue of TestMetadataTestCase.setUp) and therefore must not run in
// parallel. Method names and assertions mirror the Java; non-ASCII inputs use the
// same \u escapes as upstream. Last reconciled against: v9.0.32

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInvalidRegion(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.ZZ)
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+4", f.InputDigit('4'))
	require.Equal(t, "+48 ", f.InputDigit('8'))
	require.Equal(t, "+48 8", f.InputDigit('8'))
	require.Equal(t, "+48 88", f.InputDigit('8'))
	require.Equal(t, "+48 88 1", f.InputDigit('1'))
	require.Equal(t, "+48 88 12", f.InputDigit('2'))
	require.Equal(t, "+48 88 123", f.InputDigit('3'))
	require.Equal(t, "+48 88 123 1", f.InputDigit('1'))
	require.Equal(t, "+48 88 123 12", f.InputDigit('2'))

	f.Clear()
	require.Equal(t, "6", f.InputDigit('6'))
	require.Equal(t, "65", f.InputDigit('5'))
	require.Equal(t, "650", f.InputDigit('0'))
	require.Equal(t, "6502", f.InputDigit('2'))
	require.Equal(t, "65025", f.InputDigit('5'))
	require.Equal(t, "650253", f.InputDigit('3'))
}

func TestInvalidPlusSign(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.ZZ)
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+4", f.InputDigit('4'))
	require.Equal(t, "+48 ", f.InputDigit('8'))
	require.Equal(t, "+48 8", f.InputDigit('8'))
	require.Equal(t, "+48 88", f.InputDigit('8'))
	require.Equal(t, "+48 88 1", f.InputDigit('1'))
	require.Equal(t, "+48 88 12", f.InputDigit('2'))
	require.Equal(t, "+48 88 123", f.InputDigit('3'))
	require.Equal(t, "+48 88 123 1", f.InputDigit('1'))
	// A plus sign can only appear at the beginning of the number; otherwise, no
	// formatting is applied.
	require.Equal(t, "+48881231+", f.InputDigit('+'))
	require.Equal(t, "+48881231+2", f.InputDigit('2'))
}

func TestTooLongNumberMatchingMultipleLeadingDigits(t *testing.T) {
	useTestMetadata(t)
	// See https://github.com/google/libphonenumber/issues/36
	// The bug occurred last time for countries which have two formatting rules with
	// exactly the same leading digits pattern but differ in length.
	f := GetAsYouTypeFormatter(regionCode.ZZ)
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+81 ", f.InputDigit('1'))
	require.Equal(t, "+81 9", f.InputDigit('9'))
	require.Equal(t, "+81 90", f.InputDigit('0'))
	require.Equal(t, "+81 90 1", f.InputDigit('1'))
	require.Equal(t, "+81 90 12", f.InputDigit('2'))
	require.Equal(t, "+81 90 123", f.InputDigit('3'))
	require.Equal(t, "+81 90 1234", f.InputDigit('4'))
	require.Equal(t, "+81 90 1234 5", f.InputDigit('5'))
	require.Equal(t, "+81 90 1234 56", f.InputDigit('6'))
	require.Equal(t, "+81 90 1234 567", f.InputDigit('7'))
	require.Equal(t, "+81 90 1234 5678", f.InputDigit('8'))
	require.Equal(t, "+81 90 12 345 6789", f.InputDigit('9'))
	require.Equal(t, "+81901234567890", f.InputDigit('0'))
	require.Equal(t, "+819012345678901", f.InputDigit('1'))
}

func TestCountryWithSpaceInNationalPrefixFormattingRule(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.BY)
	require.Equal(t, "8", f.InputDigit('8'))
	require.Equal(t, "88", f.InputDigit('8'))
	require.Equal(t, "881", f.InputDigit('1'))
	require.Equal(t, "8 819", f.InputDigit('9'))
	require.Equal(t, "8 8190", f.InputDigit('0'))
	// The formatting rule for 5 digit numbers states that no space should be present
	// after the national prefix.
	require.Equal(t, "881 901", f.InputDigit('1'))
	require.Equal(t, "8 819 012", f.InputDigit('2'))
	// Too long, no formatting rule applies.
	require.Equal(t, "88190123", f.InputDigit('3'))
}

func TestCountryWithSpaceInNationalPrefixFormattingRuleAndLongNdd(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.BY)
	require.Equal(t, "9", f.InputDigit('9'))
	require.Equal(t, "99", f.InputDigit('9'))
	require.Equal(t, "999", f.InputDigit('9'))
	require.Equal(t, "9999", f.InputDigit('9'))
	require.Equal(t, "99999 ", f.InputDigit('9'))
	require.Equal(t, "99999 1", f.InputDigit('1'))
	require.Equal(t, "99999 12", f.InputDigit('2'))
	require.Equal(t, "99999 123", f.InputDigit('3'))
	require.Equal(t, "99999 1234", f.InputDigit('4'))
	require.Equal(t, "99999 12 345", f.InputDigit('5'))
}

func TestAYTFUS(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.US)
	require.Equal(t, "6", f.InputDigit('6'))
	require.Equal(t, "65", f.InputDigit('5'))
	require.Equal(t, "650", f.InputDigit('0'))
	require.Equal(t, "650 2", f.InputDigit('2'))
	require.Equal(t, "650 25", f.InputDigit('5'))
	require.Equal(t, "650 253", f.InputDigit('3'))
	// Note this is how a US local number (without area code) should be formatted.
	require.Equal(t, "650 2532", f.InputDigit('2'))
	require.Equal(t, "650 253 22", f.InputDigit('2'))
	require.Equal(t, "650 253 222", f.InputDigit('2'))
	require.Equal(t, "650 253 2222", f.InputDigit('2'))

	f.Clear()
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "16", f.InputDigit('6'))
	require.Equal(t, "1 65", f.InputDigit('5'))
	require.Equal(t, "1 650", f.InputDigit('0'))
	require.Equal(t, "1 650 2", f.InputDigit('2'))
	require.Equal(t, "1 650 25", f.InputDigit('5'))
	require.Equal(t, "1 650 253", f.InputDigit('3'))
	require.Equal(t, "1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "1 650 253 22", f.InputDigit('2'))
	require.Equal(t, "1 650 253 222", f.InputDigit('2'))
	require.Equal(t, "1 650 253 2222", f.InputDigit('2'))

	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011 ", f.InputDigit('1'))
	require.Equal(t, "011 4", f.InputDigit('4'))
	require.Equal(t, "011 44 ", f.InputDigit('4'))
	require.Equal(t, "011 44 6", f.InputDigit('6'))
	require.Equal(t, "011 44 61", f.InputDigit('1'))
	require.Equal(t, "011 44 6 12", f.InputDigit('2'))
	require.Equal(t, "011 44 6 123", f.InputDigit('3'))
	require.Equal(t, "011 44 6 123 1", f.InputDigit('1'))
	require.Equal(t, "011 44 6 123 12", f.InputDigit('2'))
	require.Equal(t, "011 44 6 123 123", f.InputDigit('3'))
	require.Equal(t, "011 44 6 123 123 1", f.InputDigit('1'))
	require.Equal(t, "011 44 6 123 123 12", f.InputDigit('2'))
	require.Equal(t, "011 44 6 123 123 123", f.InputDigit('3'))

	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011 ", f.InputDigit('1'))
	require.Equal(t, "011 5", f.InputDigit('5'))
	require.Equal(t, "011 54 ", f.InputDigit('4'))
	require.Equal(t, "011 54 9", f.InputDigit('9'))
	require.Equal(t, "011 54 91", f.InputDigit('1'))
	require.Equal(t, "011 54 9 11", f.InputDigit('1'))
	require.Equal(t, "011 54 9 11 2", f.InputDigit('2'))
	require.Equal(t, "011 54 9 11 23", f.InputDigit('3'))
	require.Equal(t, "011 54 9 11 231", f.InputDigit('1'))
	require.Equal(t, "011 54 9 11 2312", f.InputDigit('2'))
	require.Equal(t, "011 54 9 11 2312 1", f.InputDigit('1'))
	require.Equal(t, "011 54 9 11 2312 12", f.InputDigit('2'))
	require.Equal(t, "011 54 9 11 2312 123", f.InputDigit('3'))
	require.Equal(t, "011 54 9 11 2312 1234", f.InputDigit('4'))

	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011 ", f.InputDigit('1'))
	require.Equal(t, "011 2", f.InputDigit('2'))
	require.Equal(t, "011 24", f.InputDigit('4'))
	require.Equal(t, "011 244 ", f.InputDigit('4'))
	require.Equal(t, "011 244 2", f.InputDigit('2'))
	require.Equal(t, "011 244 28", f.InputDigit('8'))
	require.Equal(t, "011 244 280", f.InputDigit('0'))
	require.Equal(t, "011 244 280 0", f.InputDigit('0'))
	require.Equal(t, "011 244 280 00", f.InputDigit('0'))
	require.Equal(t, "011 244 280 000", f.InputDigit('0'))
	require.Equal(t, "011 244 280 000 0", f.InputDigit('0'))
	require.Equal(t, "011 244 280 000 00", f.InputDigit('0'))
	require.Equal(t, "011 244 280 000 000", f.InputDigit('0'))

	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+4", f.InputDigit('4'))
	require.Equal(t, "+48 ", f.InputDigit('8'))
	require.Equal(t, "+48 8", f.InputDigit('8'))
	require.Equal(t, "+48 88", f.InputDigit('8'))
	require.Equal(t, "+48 88 1", f.InputDigit('1'))
	require.Equal(t, "+48 88 12", f.InputDigit('2'))
	require.Equal(t, "+48 88 123", f.InputDigit('3'))
	require.Equal(t, "+48 88 123 1", f.InputDigit('1'))
	require.Equal(t, "+48 88 123 12", f.InputDigit('2'))
	require.Equal(t, "+48 88 123 12 1", f.InputDigit('1'))
	require.Equal(t, "+48 88 123 12 12", f.InputDigit('2'))
}

func TestAYTFUSFullWidthCharacters(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.US)
	require.Equal(t, "６", f.InputDigit('６'))
	require.Equal(t, "６５", f.InputDigit('５'))
	require.Equal(t, "650", f.InputDigit('０'))
	require.Equal(t, "650 2", f.InputDigit('２'))
	require.Equal(t, "650 25", f.InputDigit('５'))
	require.Equal(t, "650 253", f.InputDigit('３'))
	require.Equal(t, "650 2532", f.InputDigit('２'))
	require.Equal(t, "650 253 22", f.InputDigit('２'))
	require.Equal(t, "650 253 222", f.InputDigit('２'))
	require.Equal(t, "650 253 2222", f.InputDigit('２'))
}

func TestAYTFUSMobileShortCode(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.US)
	require.Equal(t, "*", f.InputDigit('*'))
	require.Equal(t, "*1", f.InputDigit('1'))
	require.Equal(t, "*12", f.InputDigit('2'))
	require.Equal(t, "*121", f.InputDigit('1'))
	require.Equal(t, "*121#", f.InputDigit('#'))
}

func TestAYTFUSVanityNumber(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.US)
	require.Equal(t, "8", f.InputDigit('8'))
	require.Equal(t, "80", f.InputDigit('0'))
	require.Equal(t, "800", f.InputDigit('0'))
	require.Equal(t, "800 ", f.InputDigit(' '))
	require.Equal(t, "800 M", f.InputDigit('M'))
	require.Equal(t, "800 MY", f.InputDigit('Y'))
	require.Equal(t, "800 MY ", f.InputDigit(' '))
	require.Equal(t, "800 MY A", f.InputDigit('A'))
	require.Equal(t, "800 MY AP", f.InputDigit('P'))
	require.Equal(t, "800 MY APP", f.InputDigit('P'))
	require.Equal(t, "800 MY APPL", f.InputDigit('L'))
	require.Equal(t, "800 MY APPLE", f.InputDigit('E'))
}

func TestAYTFAndRememberPositionUS(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.US)
	require.Equal(t, "1", f.InputDigitAndRememberPosition('1'))
	require.Equal(t, 1, f.GetRememberedPosition())
	require.Equal(t, "16", f.InputDigit('6'))
	require.Equal(t, "1 65", f.InputDigit('5'))
	require.Equal(t, 1, f.GetRememberedPosition())
	require.Equal(t, "1 650", f.InputDigitAndRememberPosition('0'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "1 650 2", f.InputDigit('2'))
	require.Equal(t, "1 650 25", f.InputDigit('5'))
	// Note the remembered position for digit "0" changes from 4 to 5, because a space
	// is now inserted in the front.
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "1 650 253", f.InputDigit('3'))
	require.Equal(t, "1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "1 650 253 22", f.InputDigit('2'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "1 650 253 222", f.InputDigitAndRememberPosition('2'))
	require.Equal(t, 13, f.GetRememberedPosition())
	require.Equal(t, "1 650 253 2222", f.InputDigit('2'))
	require.Equal(t, 13, f.GetRememberedPosition())
	require.Equal(t, "165025322222", f.InputDigit('2'))
	require.Equal(t, 10, f.GetRememberedPosition())
	require.Equal(t, "1650253222222", f.InputDigit('2'))
	require.Equal(t, 10, f.GetRememberedPosition())

	f.Clear()
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "16", f.InputDigitAndRememberPosition('6'))
	require.Equal(t, 2, f.GetRememberedPosition())
	require.Equal(t, "1 65", f.InputDigit('5'))
	require.Equal(t, "1 650", f.InputDigit('0'))
	require.Equal(t, 3, f.GetRememberedPosition())
	require.Equal(t, "1 650 2", f.InputDigit('2'))
	require.Equal(t, "1 650 25", f.InputDigit('5'))
	require.Equal(t, 3, f.GetRememberedPosition())
	require.Equal(t, "1 650 253", f.InputDigit('3'))
	require.Equal(t, "1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "1 650 253 22", f.InputDigit('2'))
	require.Equal(t, 3, f.GetRememberedPosition())
	require.Equal(t, "1 650 253 222", f.InputDigit('2'))
	require.Equal(t, "1 650 253 2222", f.InputDigit('2'))
	require.Equal(t, "165025322222", f.InputDigit('2'))
	require.Equal(t, 2, f.GetRememberedPosition())
	require.Equal(t, "1650253222222", f.InputDigit('2'))
	require.Equal(t, 2, f.GetRememberedPosition())

	f.Clear()
	require.Equal(t, "6", f.InputDigit('6'))
	require.Equal(t, "65", f.InputDigit('5'))
	require.Equal(t, "650", f.InputDigit('0'))
	require.Equal(t, "650 2", f.InputDigit('2'))
	require.Equal(t, "650 25", f.InputDigit('5'))
	require.Equal(t, "650 253", f.InputDigit('3'))
	require.Equal(t, "650 2532", f.InputDigitAndRememberPosition('2'))
	require.Equal(t, 8, f.GetRememberedPosition())
	require.Equal(t, "650 253 22", f.InputDigit('2'))
	require.Equal(t, 9, f.GetRememberedPosition())
	require.Equal(t, "650 253 222", f.InputDigit('2'))
	// No more formatting when semicolon is entered.
	require.Equal(t, "650253222;", f.InputDigit(';'))
	require.Equal(t, 7, f.GetRememberedPosition())
	require.Equal(t, "650253222;2", f.InputDigit('2'))

	f.Clear()
	require.Equal(t, "6", f.InputDigit('6'))
	require.Equal(t, "65", f.InputDigit('5'))
	require.Equal(t, "650", f.InputDigit('0'))
	// No more formatting when users choose to do their own formatting.
	require.Equal(t, "650-", f.InputDigit('-'))
	require.Equal(t, "650-2", f.InputDigitAndRememberPosition('2'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "650-25", f.InputDigit('5'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "650-253", f.InputDigit('3'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "650-253-", f.InputDigit('-'))
	require.Equal(t, "650-253-2", f.InputDigit('2'))
	require.Equal(t, "650-253-22", f.InputDigit('2'))
	require.Equal(t, "650-253-222", f.InputDigit('2'))
	require.Equal(t, "650-253-2222", f.InputDigit('2'))

	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011 ", f.InputDigit('1'))
	require.Equal(t, "011 4", f.InputDigitAndRememberPosition('4'))
	require.Equal(t, "011 48 ", f.InputDigit('8'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "011 48 8", f.InputDigit('8'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "011 48 88", f.InputDigit('8'))
	require.Equal(t, "011 48 88 1", f.InputDigit('1'))
	require.Equal(t, "011 48 88 12", f.InputDigit('2'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "011 48 88 123", f.InputDigit('3'))
	require.Equal(t, "011 48 88 123 1", f.InputDigit('1'))
	require.Equal(t, "011 48 88 123 12", f.InputDigit('2'))
	require.Equal(t, "011 48 88 123 12 1", f.InputDigit('1'))
	require.Equal(t, "011 48 88 123 12 12", f.InputDigit('2'))

	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+1", f.InputDigit('1'))
	require.Equal(t, "+1 6", f.InputDigitAndRememberPosition('6'))
	require.Equal(t, "+1 65", f.InputDigit('5'))
	require.Equal(t, "+1 650", f.InputDigit('0'))
	require.Equal(t, 4, f.GetRememberedPosition())
	require.Equal(t, "+1 650 2", f.InputDigit('2'))
	require.Equal(t, 4, f.GetRememberedPosition())
	require.Equal(t, "+1 650 25", f.InputDigit('5'))
	require.Equal(t, "+1 650 253", f.InputDigitAndRememberPosition('3'))
	require.Equal(t, "+1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "+1 650 253 22", f.InputDigit('2'))
	require.Equal(t, "+1 650 253 222", f.InputDigit('2'))
	require.Equal(t, 10, f.GetRememberedPosition())

	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+1", f.InputDigit('1'))
	require.Equal(t, "+1 6", f.InputDigitAndRememberPosition('6'))
	require.Equal(t, "+1 65", f.InputDigit('5'))
	require.Equal(t, "+1 650", f.InputDigit('0'))
	require.Equal(t, 4, f.GetRememberedPosition())
	require.Equal(t, "+1 650 2", f.InputDigit('2'))
	require.Equal(t, 4, f.GetRememberedPosition())
	require.Equal(t, "+1 650 25", f.InputDigit('5'))
	require.Equal(t, "+1 650 253", f.InputDigit('3'))
	require.Equal(t, "+1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "+1 650 253 22", f.InputDigit('2'))
	require.Equal(t, "+1 650 253 222", f.InputDigit('2'))
	require.Equal(t, "+1650253222;", f.InputDigit(';'))
	require.Equal(t, 3, f.GetRememberedPosition())
}

func TestAYTFGBFixedLine(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.GB)
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "02", f.InputDigit('2'))
	require.Equal(t, "020", f.InputDigit('0'))
	require.Equal(t, "020 7", f.InputDigitAndRememberPosition('7'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "020 70", f.InputDigit('0'))
	require.Equal(t, "020 703", f.InputDigit('3'))
	require.Equal(t, 5, f.GetRememberedPosition())
	require.Equal(t, "020 7031", f.InputDigit('1'))
	require.Equal(t, "020 7031 3", f.InputDigit('3'))
	require.Equal(t, "020 7031 30", f.InputDigit('0'))
	require.Equal(t, "020 7031 300", f.InputDigit('0'))
	require.Equal(t, "020 7031 3000", f.InputDigit('0'))
}

func TestAYTFGBTollFree(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.GB)
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "08", f.InputDigit('8'))
	require.Equal(t, "080", f.InputDigit('0'))
	require.Equal(t, "080 7", f.InputDigit('7'))
	require.Equal(t, "080 70", f.InputDigit('0'))
	require.Equal(t, "080 703", f.InputDigit('3'))
	require.Equal(t, "080 7031", f.InputDigit('1'))
	require.Equal(t, "080 7031 3", f.InputDigit('3'))
	require.Equal(t, "080 7031 30", f.InputDigit('0'))
	require.Equal(t, "080 7031 300", f.InputDigit('0'))
	require.Equal(t, "080 7031 3000", f.InputDigit('0'))
}

func TestAYTFGBPremiumRate(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.GB)
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "09", f.InputDigit('9'))
	require.Equal(t, "090", f.InputDigit('0'))
	require.Equal(t, "090 7", f.InputDigit('7'))
	require.Equal(t, "090 70", f.InputDigit('0'))
	require.Equal(t, "090 703", f.InputDigit('3'))
	require.Equal(t, "090 7031", f.InputDigit('1'))
	require.Equal(t, "090 7031 3", f.InputDigit('3'))
	require.Equal(t, "090 7031 30", f.InputDigit('0'))
	require.Equal(t, "090 7031 300", f.InputDigit('0'))
	require.Equal(t, "090 7031 3000", f.InputDigit('0'))
}

func TestAYTFNZMobile(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.NZ)
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "02", f.InputDigit('2'))
	require.Equal(t, "021", f.InputDigit('1'))
	require.Equal(t, "02-11", f.InputDigit('1'))
	require.Equal(t, "02-112", f.InputDigit('2'))
	// Note the unittest is using fake metadata which might produce non-ideal results.
	require.Equal(t, "02-112 3", f.InputDigit('3'))
	require.Equal(t, "02-112 34", f.InputDigit('4'))
	require.Equal(t, "02-112 345", f.InputDigit('5'))
	require.Equal(t, "02-112 3456", f.InputDigit('6'))
}

func TestAYTFDE(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.DE)
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "03", f.InputDigit('3'))
	require.Equal(t, "030", f.InputDigit('0'))
	require.Equal(t, "030/1", f.InputDigit('1'))
	require.Equal(t, "030/12", f.InputDigit('2'))
	require.Equal(t, "030/123", f.InputDigit('3'))
	require.Equal(t, "030/1234", f.InputDigit('4'))

	// 04134 1234
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "04", f.InputDigit('4'))
	require.Equal(t, "041", f.InputDigit('1'))
	require.Equal(t, "041 3", f.InputDigit('3'))
	require.Equal(t, "041 34", f.InputDigit('4'))
	require.Equal(t, "04134 1", f.InputDigit('1'))
	require.Equal(t, "04134 12", f.InputDigit('2'))
	require.Equal(t, "04134 123", f.InputDigit('3'))
	require.Equal(t, "04134 1234", f.InputDigit('4'))

	// 08021 2345
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "08", f.InputDigit('8'))
	require.Equal(t, "080", f.InputDigit('0'))
	require.Equal(t, "080 2", f.InputDigit('2'))
	require.Equal(t, "080 21", f.InputDigit('1'))
	require.Equal(t, "08021 2", f.InputDigit('2'))
	require.Equal(t, "08021 23", f.InputDigit('3'))
	require.Equal(t, "08021 234", f.InputDigit('4'))
	require.Equal(t, "08021 2345", f.InputDigit('5'))

	// 00 1 650 253 2250
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "00", f.InputDigit('0'))
	require.Equal(t, "00 1 ", f.InputDigit('1'))
	require.Equal(t, "00 1 6", f.InputDigit('6'))
	require.Equal(t, "00 1 65", f.InputDigit('5'))
	require.Equal(t, "00 1 650", f.InputDigit('0'))
	require.Equal(t, "00 1 650 2", f.InputDigit('2'))
	require.Equal(t, "00 1 650 25", f.InputDigit('5'))
	require.Equal(t, "00 1 650 253", f.InputDigit('3'))
	require.Equal(t, "00 1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "00 1 650 253 22", f.InputDigit('2'))
	require.Equal(t, "00 1 650 253 222", f.InputDigit('2'))
	require.Equal(t, "00 1 650 253 2222", f.InputDigit('2'))
}

func TestAYTFAR(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.AR)
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011", f.InputDigit('1'))
	require.Equal(t, "011 7", f.InputDigit('7'))
	require.Equal(t, "011 70", f.InputDigit('0'))
	require.Equal(t, "011 703", f.InputDigit('3'))
	require.Equal(t, "011 7031", f.InputDigit('1'))
	require.Equal(t, "011 7031-3", f.InputDigit('3'))
	require.Equal(t, "011 7031-30", f.InputDigit('0'))
	require.Equal(t, "011 7031-300", f.InputDigit('0'))
	require.Equal(t, "011 7031-3000", f.InputDigit('0'))
}

func TestAYTFARMobile(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.AR)
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+5", f.InputDigit('5'))
	require.Equal(t, "+54 ", f.InputDigit('4'))
	require.Equal(t, "+54 9", f.InputDigit('9'))
	require.Equal(t, "+54 91", f.InputDigit('1'))
	require.Equal(t, "+54 9 11", f.InputDigit('1'))
	require.Equal(t, "+54 9 11 2", f.InputDigit('2'))
	require.Equal(t, "+54 9 11 23", f.InputDigit('3'))
	require.Equal(t, "+54 9 11 231", f.InputDigit('1'))
	require.Equal(t, "+54 9 11 2312", f.InputDigit('2'))
	require.Equal(t, "+54 9 11 2312 1", f.InputDigit('1'))
	require.Equal(t, "+54 9 11 2312 12", f.InputDigit('2'))
	require.Equal(t, "+54 9 11 2312 123", f.InputDigit('3'))
	require.Equal(t, "+54 9 11 2312 1234", f.InputDigit('4'))
}

func TestAYTFKR(t *testing.T) {
	useTestMetadata(t)
	// +82 51 234 5678
	f := GetAsYouTypeFormatter(regionCode.KR)
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+82 ", f.InputDigit('2'))
	require.Equal(t, "+82 5", f.InputDigit('5'))
	require.Equal(t, "+82 51", f.InputDigit('1'))
	require.Equal(t, "+82 51-2", f.InputDigit('2'))
	require.Equal(t, "+82 51-23", f.InputDigit('3'))
	require.Equal(t, "+82 51-234", f.InputDigit('4'))
	require.Equal(t, "+82 51-234-5", f.InputDigit('5'))
	require.Equal(t, "+82 51-234-56", f.InputDigit('6'))
	require.Equal(t, "+82 51-234-567", f.InputDigit('7'))
	require.Equal(t, "+82 51-234-5678", f.InputDigit('8'))

	// +82 2 531 5678
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+82 ", f.InputDigit('2'))
	require.Equal(t, "+82 2", f.InputDigit('2'))
	require.Equal(t, "+82 25", f.InputDigit('5'))
	require.Equal(t, "+82 2-53", f.InputDigit('3'))
	require.Equal(t, "+82 2-531", f.InputDigit('1'))
	require.Equal(t, "+82 2-531-5", f.InputDigit('5'))
	require.Equal(t, "+82 2-531-56", f.InputDigit('6'))
	require.Equal(t, "+82 2-531-567", f.InputDigit('7'))
	require.Equal(t, "+82 2-531-5678", f.InputDigit('8'))

	// +82 2 3665 5678
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+82 ", f.InputDigit('2'))
	require.Equal(t, "+82 2", f.InputDigit('2'))
	require.Equal(t, "+82 23", f.InputDigit('3'))
	require.Equal(t, "+82 2-36", f.InputDigit('6'))
	require.Equal(t, "+82 2-366", f.InputDigit('6'))
	require.Equal(t, "+82 2-3665", f.InputDigit('5'))
	require.Equal(t, "+82 2-3665-5", f.InputDigit('5'))
	require.Equal(t, "+82 2-3665-56", f.InputDigit('6'))
	require.Equal(t, "+82 2-3665-567", f.InputDigit('7'))
	require.Equal(t, "+82 2-3665-5678", f.InputDigit('8'))

	// 02-114
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "02", f.InputDigit('2'))
	require.Equal(t, "021", f.InputDigit('1'))
	require.Equal(t, "02-11", f.InputDigit('1'))
	require.Equal(t, "02-114", f.InputDigit('4'))

	// 02-1300
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "02", f.InputDigit('2'))
	require.Equal(t, "021", f.InputDigit('1'))
	require.Equal(t, "02-13", f.InputDigit('3'))
	require.Equal(t, "02-130", f.InputDigit('0'))
	require.Equal(t, "02-1300", f.InputDigit('0'))

	// 011-456-7890
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011", f.InputDigit('1'))
	require.Equal(t, "011-4", f.InputDigit('4'))
	require.Equal(t, "011-45", f.InputDigit('5'))
	require.Equal(t, "011-456", f.InputDigit('6'))
	require.Equal(t, "011-456-7", f.InputDigit('7'))
	require.Equal(t, "011-456-78", f.InputDigit('8'))
	require.Equal(t, "011-456-789", f.InputDigit('9'))
	require.Equal(t, "011-456-7890", f.InputDigit('0'))

	// 011-9876-7890
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011", f.InputDigit('1'))
	require.Equal(t, "011-9", f.InputDigit('9'))
	require.Equal(t, "011-98", f.InputDigit('8'))
	require.Equal(t, "011-987", f.InputDigit('7'))
	require.Equal(t, "011-9876", f.InputDigit('6'))
	require.Equal(t, "011-9876-7", f.InputDigit('7'))
	require.Equal(t, "011-9876-78", f.InputDigit('8'))
	require.Equal(t, "011-9876-789", f.InputDigit('9'))
	require.Equal(t, "011-9876-7890", f.InputDigit('0'))
}

func TestAYTF_MX(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.MX)

	// +52 800 123 4567
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+5", f.InputDigit('5'))
	require.Equal(t, "+52 ", f.InputDigit('2'))
	require.Equal(t, "+52 8", f.InputDigit('8'))
	require.Equal(t, "+52 80", f.InputDigit('0'))
	require.Equal(t, "+52 800", f.InputDigit('0'))
	require.Equal(t, "+52 800 1", f.InputDigit('1'))
	require.Equal(t, "+52 800 12", f.InputDigit('2'))
	require.Equal(t, "+52 800 123", f.InputDigit('3'))
	require.Equal(t, "+52 800 123 4", f.InputDigit('4'))
	require.Equal(t, "+52 800 123 45", f.InputDigit('5'))
	require.Equal(t, "+52 800 123 456", f.InputDigit('6'))
	require.Equal(t, "+52 800 123 4567", f.InputDigit('7'))

	// +529011234567, proactively ensuring that no formatting is applied, where a
	// format is chosen that would otherwise have led to some digits being dropped.
	f.Clear()
	require.Equal(t, "9", f.InputDigit('9'))
	require.Equal(t, "90", f.InputDigit('0'))
	require.Equal(t, "901", f.InputDigit('1'))
	require.Equal(t, "9011", f.InputDigit('1'))
	require.Equal(t, "90112", f.InputDigit('2'))
	require.Equal(t, "901123", f.InputDigit('3'))
	require.Equal(t, "9011234", f.InputDigit('4'))
	require.Equal(t, "90112345", f.InputDigit('5'))
	require.Equal(t, "901123456", f.InputDigit('6'))
	require.Equal(t, "9011234567", f.InputDigit('7'))

	// +52 55 1234 5678
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+5", f.InputDigit('5'))
	require.Equal(t, "+52 ", f.InputDigit('2'))
	require.Equal(t, "+52 5", f.InputDigit('5'))
	require.Equal(t, "+52 55", f.InputDigit('5'))
	require.Equal(t, "+52 55 1", f.InputDigit('1'))
	require.Equal(t, "+52 55 12", f.InputDigit('2'))
	require.Equal(t, "+52 55 123", f.InputDigit('3'))
	require.Equal(t, "+52 55 1234", f.InputDigit('4'))
	require.Equal(t, "+52 55 1234 5", f.InputDigit('5'))
	require.Equal(t, "+52 55 1234 56", f.InputDigit('6'))
	require.Equal(t, "+52 55 1234 567", f.InputDigit('7'))
	require.Equal(t, "+52 55 1234 5678", f.InputDigit('8'))

	// +52 212 345 6789
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+5", f.InputDigit('5'))
	require.Equal(t, "+52 ", f.InputDigit('2'))
	require.Equal(t, "+52 2", f.InputDigit('2'))
	require.Equal(t, "+52 21", f.InputDigit('1'))
	require.Equal(t, "+52 212", f.InputDigit('2'))
	require.Equal(t, "+52 212 3", f.InputDigit('3'))
	require.Equal(t, "+52 212 34", f.InputDigit('4'))
	require.Equal(t, "+52 212 345", f.InputDigit('5'))
	require.Equal(t, "+52 212 345 6", f.InputDigit('6'))
	require.Equal(t, "+52 212 345 67", f.InputDigit('7'))
	require.Equal(t, "+52 212 345 678", f.InputDigit('8'))
	require.Equal(t, "+52 212 345 6789", f.InputDigit('9'))

	// +52 1 55 1234 5678
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+5", f.InputDigit('5'))
	require.Equal(t, "+52 ", f.InputDigit('2'))
	require.Equal(t, "+52 1", f.InputDigit('1'))
	require.Equal(t, "+52 15", f.InputDigit('5'))
	require.Equal(t, "+52 1 55", f.InputDigit('5'))
	require.Equal(t, "+52 1 55 1", f.InputDigit('1'))
	require.Equal(t, "+52 1 55 12", f.InputDigit('2'))
	require.Equal(t, "+52 1 55 123", f.InputDigit('3'))
	require.Equal(t, "+52 1 55 1234", f.InputDigit('4'))
	require.Equal(t, "+52 1 55 1234 5", f.InputDigit('5'))
	require.Equal(t, "+52 1 55 1234 56", f.InputDigit('6'))
	require.Equal(t, "+52 1 55 1234 567", f.InputDigit('7'))
	require.Equal(t, "+52 1 55 1234 5678", f.InputDigit('8'))

	// +52 1 541 234 5678
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+5", f.InputDigit('5'))
	require.Equal(t, "+52 ", f.InputDigit('2'))
	require.Equal(t, "+52 1", f.InputDigit('1'))
	require.Equal(t, "+52 15", f.InputDigit('5'))
	require.Equal(t, "+52 1 54", f.InputDigit('4'))
	require.Equal(t, "+52 1 541", f.InputDigit('1'))
	require.Equal(t, "+52 1 541 2", f.InputDigit('2'))
	require.Equal(t, "+52 1 541 23", f.InputDigit('3'))
	require.Equal(t, "+52 1 541 234", f.InputDigit('4'))
	require.Equal(t, "+52 1 541 234 5", f.InputDigit('5'))
	require.Equal(t, "+52 1 541 234 56", f.InputDigit('6'))
	require.Equal(t, "+52 1 541 234 567", f.InputDigit('7'))
	require.Equal(t, "+52 1 541 234 5678", f.InputDigit('8'))
}

func TestAYTF_International_Toll_Free(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.US)
	// +800 1234 5678
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+80", f.InputDigit('0'))
	require.Equal(t, "+800 ", f.InputDigit('0'))
	require.Equal(t, "+800 1", f.InputDigit('1'))
	require.Equal(t, "+800 12", f.InputDigit('2'))
	require.Equal(t, "+800 123", f.InputDigit('3'))
	require.Equal(t, "+800 1234", f.InputDigit('4'))
	require.Equal(t, "+800 1234 5", f.InputDigit('5'))
	require.Equal(t, "+800 1234 56", f.InputDigit('6'))
	require.Equal(t, "+800 1234 567", f.InputDigit('7'))
	require.Equal(t, "+800 1234 5678", f.InputDigit('8'))
	require.Equal(t, "+800123456789", f.InputDigit('9'))
}

func TestAYTFMultipleLeadingDigitPatterns(t *testing.T) {
	useTestMetadata(t)
	// +81 50 2345 6789
	f := GetAsYouTypeFormatter(regionCode.JP)
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+81 ", f.InputDigit('1'))
	require.Equal(t, "+81 5", f.InputDigit('5'))
	require.Equal(t, "+81 50", f.InputDigit('0'))
	require.Equal(t, "+81 50 2", f.InputDigit('2'))
	require.Equal(t, "+81 50 23", f.InputDigit('3'))
	require.Equal(t, "+81 50 234", f.InputDigit('4'))
	require.Equal(t, "+81 50 2345", f.InputDigit('5'))
	require.Equal(t, "+81 50 2345 6", f.InputDigit('6'))
	require.Equal(t, "+81 50 2345 67", f.InputDigit('7'))
	require.Equal(t, "+81 50 2345 678", f.InputDigit('8'))
	require.Equal(t, "+81 50 2345 6789", f.InputDigit('9'))

	// +81 222 12 5678
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+81 ", f.InputDigit('1'))
	require.Equal(t, "+81 2", f.InputDigit('2'))
	require.Equal(t, "+81 22", f.InputDigit('2'))
	require.Equal(t, "+81 22 2", f.InputDigit('2'))
	require.Equal(t, "+81 22 21", f.InputDigit('1'))
	require.Equal(t, "+81 2221 2", f.InputDigit('2'))
	require.Equal(t, "+81 222 12 5", f.InputDigit('5'))
	require.Equal(t, "+81 222 12 56", f.InputDigit('6'))
	require.Equal(t, "+81 222 12 567", f.InputDigit('7'))
	require.Equal(t, "+81 222 12 5678", f.InputDigit('8'))

	// 011113
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "01", f.InputDigit('1'))
	require.Equal(t, "011", f.InputDigit('1'))
	require.Equal(t, "011 1", f.InputDigit('1'))
	require.Equal(t, "011 11", f.InputDigit('1'))
	require.Equal(t, "011113", f.InputDigit('3'))

	// +81 3332 2 5678
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+81 ", f.InputDigit('1'))
	require.Equal(t, "+81 3", f.InputDigit('3'))
	require.Equal(t, "+81 33", f.InputDigit('3'))
	require.Equal(t, "+81 33 3", f.InputDigit('3'))
	require.Equal(t, "+81 3332", f.InputDigit('2'))
	require.Equal(t, "+81 3332 2", f.InputDigit('2'))
	require.Equal(t, "+81 3332 2 5", f.InputDigit('5'))
	require.Equal(t, "+81 3332 2 56", f.InputDigit('6'))
	require.Equal(t, "+81 3332 2 567", f.InputDigit('7'))
	require.Equal(t, "+81 3332 2 5678", f.InputDigit('8'))
}

func TestAYTFLongIDD_AU(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.AU)
	// 0011 1 650 253 2250
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "00", f.InputDigit('0'))
	require.Equal(t, "001", f.InputDigit('1'))
	require.Equal(t, "0011", f.InputDigit('1'))
	require.Equal(t, "0011 1 ", f.InputDigit('1'))
	require.Equal(t, "0011 1 6", f.InputDigit('6'))
	require.Equal(t, "0011 1 65", f.InputDigit('5'))
	require.Equal(t, "0011 1 650", f.InputDigit('0'))
	require.Equal(t, "0011 1 650 2", f.InputDigit('2'))
	require.Equal(t, "0011 1 650 25", f.InputDigit('5'))
	require.Equal(t, "0011 1 650 253", f.InputDigit('3'))
	require.Equal(t, "0011 1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "0011 1 650 253 22", f.InputDigit('2'))
	require.Equal(t, "0011 1 650 253 222", f.InputDigit('2'))
	require.Equal(t, "0011 1 650 253 2222", f.InputDigit('2'))

	// 0011 81 3332 2 5678
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "00", f.InputDigit('0'))
	require.Equal(t, "001", f.InputDigit('1'))
	require.Equal(t, "0011", f.InputDigit('1'))
	require.Equal(t, "00118", f.InputDigit('8'))
	require.Equal(t, "0011 81 ", f.InputDigit('1'))
	require.Equal(t, "0011 81 3", f.InputDigit('3'))
	require.Equal(t, "0011 81 33", f.InputDigit('3'))
	require.Equal(t, "0011 81 33 3", f.InputDigit('3'))
	require.Equal(t, "0011 81 3332", f.InputDigit('2'))
	require.Equal(t, "0011 81 3332 2", f.InputDigit('2'))
	require.Equal(t, "0011 81 3332 2 5", f.InputDigit('5'))
	require.Equal(t, "0011 81 3332 2 56", f.InputDigit('6'))
	require.Equal(t, "0011 81 3332 2 567", f.InputDigit('7'))
	require.Equal(t, "0011 81 3332 2 5678", f.InputDigit('8'))

	// 0011 244 250 253 222
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "00", f.InputDigit('0'))
	require.Equal(t, "001", f.InputDigit('1'))
	require.Equal(t, "0011", f.InputDigit('1'))
	require.Equal(t, "00112", f.InputDigit('2'))
	require.Equal(t, "001124", f.InputDigit('4'))
	require.Equal(t, "0011 244 ", f.InputDigit('4'))
	require.Equal(t, "0011 244 2", f.InputDigit('2'))
	require.Equal(t, "0011 244 25", f.InputDigit('5'))
	require.Equal(t, "0011 244 250", f.InputDigit('0'))
	require.Equal(t, "0011 244 250 2", f.InputDigit('2'))
	require.Equal(t, "0011 244 250 25", f.InputDigit('5'))
	require.Equal(t, "0011 244 250 253", f.InputDigit('3'))
	require.Equal(t, "0011 244 250 253 2", f.InputDigit('2'))
	require.Equal(t, "0011 244 250 253 22", f.InputDigit('2'))
	require.Equal(t, "0011 244 250 253 222", f.InputDigit('2'))
}

func TestAYTFLongIDD_KR(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.KR)
	// 00300 1 650 253 2222
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "00", f.InputDigit('0'))
	require.Equal(t, "003", f.InputDigit('3'))
	require.Equal(t, "0030", f.InputDigit('0'))
	require.Equal(t, "00300", f.InputDigit('0'))
	require.Equal(t, "00300 1 ", f.InputDigit('1'))
	require.Equal(t, "00300 1 6", f.InputDigit('6'))
	require.Equal(t, "00300 1 65", f.InputDigit('5'))
	require.Equal(t, "00300 1 650", f.InputDigit('0'))
	require.Equal(t, "00300 1 650 2", f.InputDigit('2'))
	require.Equal(t, "00300 1 650 25", f.InputDigit('5'))
	require.Equal(t, "00300 1 650 253", f.InputDigit('3'))
	require.Equal(t, "00300 1 650 253 2", f.InputDigit('2'))
	require.Equal(t, "00300 1 650 253 22", f.InputDigit('2'))
	require.Equal(t, "00300 1 650 253 222", f.InputDigit('2'))
	require.Equal(t, "00300 1 650 253 2222", f.InputDigit('2'))
}

func TestAYTFLongNDD_KR(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.KR)
	// 08811-9876-7890
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "08", f.InputDigit('8'))
	require.Equal(t, "088", f.InputDigit('8'))
	require.Equal(t, "0881", f.InputDigit('1'))
	require.Equal(t, "08811", f.InputDigit('1'))
	require.Equal(t, "08811-9", f.InputDigit('9'))
	require.Equal(t, "08811-98", f.InputDigit('8'))
	require.Equal(t, "08811-987", f.InputDigit('7'))
	require.Equal(t, "08811-9876", f.InputDigit('6'))
	require.Equal(t, "08811-9876-7", f.InputDigit('7'))
	require.Equal(t, "08811-9876-78", f.InputDigit('8'))
	require.Equal(t, "08811-9876-789", f.InputDigit('9'))
	require.Equal(t, "08811-9876-7890", f.InputDigit('0'))

	// 08500 11-9876-7890
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "08", f.InputDigit('8'))
	require.Equal(t, "085", f.InputDigit('5'))
	require.Equal(t, "0850", f.InputDigit('0'))
	require.Equal(t, "08500 ", f.InputDigit('0'))
	require.Equal(t, "08500 1", f.InputDigit('1'))
	require.Equal(t, "08500 11", f.InputDigit('1'))
	require.Equal(t, "08500 11-9", f.InputDigit('9'))
	require.Equal(t, "08500 11-98", f.InputDigit('8'))
	require.Equal(t, "08500 11-987", f.InputDigit('7'))
	require.Equal(t, "08500 11-9876", f.InputDigit('6'))
	require.Equal(t, "08500 11-9876-7", f.InputDigit('7'))
	require.Equal(t, "08500 11-9876-78", f.InputDigit('8'))
	require.Equal(t, "08500 11-9876-789", f.InputDigit('9'))
	require.Equal(t, "08500 11-9876-7890", f.InputDigit('0'))
}

func TestAYTFLongNDD_SG(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.SG)
	// 777777 9876 7890
	require.Equal(t, "7", f.InputDigit('7'))
	require.Equal(t, "77", f.InputDigit('7'))
	require.Equal(t, "777", f.InputDigit('7'))
	require.Equal(t, "7777", f.InputDigit('7'))
	require.Equal(t, "77777", f.InputDigit('7'))
	require.Equal(t, "777777 ", f.InputDigit('7'))
	require.Equal(t, "777777 9", f.InputDigit('9'))
	require.Equal(t, "777777 98", f.InputDigit('8'))
	require.Equal(t, "777777 987", f.InputDigit('7'))
	require.Equal(t, "777777 9876", f.InputDigit('6'))
	require.Equal(t, "777777 9876 7", f.InputDigit('7'))
	require.Equal(t, "777777 9876 78", f.InputDigit('8'))
	require.Equal(t, "777777 9876 789", f.InputDigit('9'))
	require.Equal(t, "777777 9876 7890", f.InputDigit('0'))
}

func TestAYTFShortNumberFormattingFix_AU(t *testing.T) {
	useTestMetadata(t)
	// For Australia, the national prefix is not optional when formatting.
	f := GetAsYouTypeFormatter(regionCode.AU)

	// 1234567890 - For leading digit 1, the national prefix formatting rule has first
	// group only.
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "12", f.InputDigit('2'))
	require.Equal(t, "123", f.InputDigit('3'))
	require.Equal(t, "1234", f.InputDigit('4'))
	require.Equal(t, "1234 5", f.InputDigit('5'))
	require.Equal(t, "1234 56", f.InputDigit('6'))
	require.Equal(t, "1234 567", f.InputDigit('7'))
	require.Equal(t, "1234 567 8", f.InputDigit('8'))
	require.Equal(t, "1234 567 89", f.InputDigit('9'))
	require.Equal(t, "1234 567 890", f.InputDigit('0'))

	// +61 1234 567 890 - Test the same number, but with the country code.
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+6", f.InputDigit('6'))
	require.Equal(t, "+61 ", f.InputDigit('1'))
	require.Equal(t, "+61 1", f.InputDigit('1'))
	require.Equal(t, "+61 12", f.InputDigit('2'))
	require.Equal(t, "+61 123", f.InputDigit('3'))
	require.Equal(t, "+61 1234", f.InputDigit('4'))
	require.Equal(t, "+61 1234 5", f.InputDigit('5'))
	require.Equal(t, "+61 1234 56", f.InputDigit('6'))
	require.Equal(t, "+61 1234 567", f.InputDigit('7'))
	require.Equal(t, "+61 1234 567 8", f.InputDigit('8'))
	require.Equal(t, "+61 1234 567 89", f.InputDigit('9'))
	require.Equal(t, "+61 1234 567 890", f.InputDigit('0'))

	// 212345678 - For leading digit 2, the national prefix formatting rule puts the
	// national prefix before the first group.
	f.Clear()
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "02", f.InputDigit('2'))
	require.Equal(t, "021", f.InputDigit('1'))
	require.Equal(t, "02 12", f.InputDigit('2'))
	require.Equal(t, "02 123", f.InputDigit('3'))
	require.Equal(t, "02 1234", f.InputDigit('4'))
	require.Equal(t, "02 1234 5", f.InputDigit('5'))
	require.Equal(t, "02 1234 56", f.InputDigit('6'))
	require.Equal(t, "02 1234 567", f.InputDigit('7'))
	require.Equal(t, "02 1234 5678", f.InputDigit('8'))

	// 212345678 - Test the same number, but without the leading 0.
	f.Clear()
	require.Equal(t, "2", f.InputDigit('2'))
	require.Equal(t, "21", f.InputDigit('1'))
	require.Equal(t, "212", f.InputDigit('2'))
	require.Equal(t, "2123", f.InputDigit('3'))
	require.Equal(t, "21234", f.InputDigit('4'))
	require.Equal(t, "212345", f.InputDigit('5'))
	require.Equal(t, "2123456", f.InputDigit('6'))
	require.Equal(t, "21234567", f.InputDigit('7'))
	require.Equal(t, "212345678", f.InputDigit('8'))

	// +61 2 1234 5678 - Test the same number, but with the country code.
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+6", f.InputDigit('6'))
	require.Equal(t, "+61 ", f.InputDigit('1'))
	require.Equal(t, "+61 2", f.InputDigit('2'))
	require.Equal(t, "+61 21", f.InputDigit('1'))
	require.Equal(t, "+61 2 12", f.InputDigit('2'))
	require.Equal(t, "+61 2 123", f.InputDigit('3'))
	require.Equal(t, "+61 2 1234", f.InputDigit('4'))
	require.Equal(t, "+61 2 1234 5", f.InputDigit('5'))
	require.Equal(t, "+61 2 1234 56", f.InputDigit('6'))
	require.Equal(t, "+61 2 1234 567", f.InputDigit('7'))
	require.Equal(t, "+61 2 1234 5678", f.InputDigit('8'))
}

func TestAYTFShortNumberFormattingFix_KR(t *testing.T) {
	useTestMetadata(t)
	// For Korea, the national prefix is not optional when formatting, and the national
	// prefix formatting rule doesn't consist of only the first group.
	f := GetAsYouTypeFormatter(regionCode.KR)

	// 111
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "11", f.InputDigit('1'))
	require.Equal(t, "111", f.InputDigit('1'))

	// 114
	f.Clear()
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "11", f.InputDigit('1'))
	require.Equal(t, "114", f.InputDigit('4'))

	// 13121234 - Test a mobile number without the national prefix. Even though it is
	// not an emergency number, it should be formatted as a block.
	f.Clear()
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "13", f.InputDigit('3'))
	require.Equal(t, "131", f.InputDigit('1'))
	require.Equal(t, "1312", f.InputDigit('2'))
	require.Equal(t, "13121", f.InputDigit('1'))
	require.Equal(t, "131212", f.InputDigit('2'))
	require.Equal(t, "1312123", f.InputDigit('3'))
	require.Equal(t, "13121234", f.InputDigit('4'))

	// +82 131-2-1234 - Test the same number, but with the country code.
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+82 ", f.InputDigit('2'))
	require.Equal(t, "+82 1", f.InputDigit('1'))
	require.Equal(t, "+82 13", f.InputDigit('3'))
	require.Equal(t, "+82 131", f.InputDigit('1'))
	require.Equal(t, "+82 131-2", f.InputDigit('2'))
	require.Equal(t, "+82 131-2-1", f.InputDigit('1'))
	require.Equal(t, "+82 131-2-12", f.InputDigit('2'))
	require.Equal(t, "+82 131-2-123", f.InputDigit('3'))
	require.Equal(t, "+82 131-2-1234", f.InputDigit('4'))
}

func TestAYTFShortNumberFormattingFix_MX(t *testing.T) {
	useTestMetadata(t)
	// For Mexico, the national prefix is optional when formatting.
	f := GetAsYouTypeFormatter(regionCode.MX)

	// 911
	require.Equal(t, "9", f.InputDigit('9'))
	require.Equal(t, "91", f.InputDigit('1'))
	require.Equal(t, "911", f.InputDigit('1'))

	// 800 123 4567 - Test a toll-free number, which should have a formatting rule
	// applied to it even though it doesn't begin with the national prefix.
	f.Clear()
	require.Equal(t, "8", f.InputDigit('8'))
	require.Equal(t, "80", f.InputDigit('0'))
	require.Equal(t, "800", f.InputDigit('0'))
	require.Equal(t, "800 1", f.InputDigit('1'))
	require.Equal(t, "800 12", f.InputDigit('2'))
	require.Equal(t, "800 123", f.InputDigit('3'))
	require.Equal(t, "800 123 4", f.InputDigit('4'))
	require.Equal(t, "800 123 45", f.InputDigit('5'))
	require.Equal(t, "800 123 456", f.InputDigit('6'))
	require.Equal(t, "800 123 4567", f.InputDigit('7'))

	// +52 800 123 4567 - Test the same number, but with the country code.
	f.Clear()
	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+5", f.InputDigit('5'))
	require.Equal(t, "+52 ", f.InputDigit('2'))
	require.Equal(t, "+52 8", f.InputDigit('8'))
	require.Equal(t, "+52 80", f.InputDigit('0'))
	require.Equal(t, "+52 800", f.InputDigit('0'))
	require.Equal(t, "+52 800 1", f.InputDigit('1'))
	require.Equal(t, "+52 800 12", f.InputDigit('2'))
	require.Equal(t, "+52 800 123", f.InputDigit('3'))
	require.Equal(t, "+52 800 123 4", f.InputDigit('4'))
	require.Equal(t, "+52 800 123 45", f.InputDigit('5'))
	require.Equal(t, "+52 800 123 456", f.InputDigit('6'))
	require.Equal(t, "+52 800 123 4567", f.InputDigit('7'))
}

func TestAYTFNoNationalPrefix(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.IT)

	require.Equal(t, "3", f.InputDigit('3'))
	require.Equal(t, "33", f.InputDigit('3'))
	require.Equal(t, "333", f.InputDigit('3'))
	require.Equal(t, "333 3", f.InputDigit('3'))
	require.Equal(t, "333 33", f.InputDigit('3'))
	require.Equal(t, "333 333", f.InputDigit('3'))
}

func TestAYTFNoNationalPrefixFormattingRule(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.AO)

	require.Equal(t, "3", f.InputDigit('3'))
	require.Equal(t, "33", f.InputDigit('3'))
	require.Equal(t, "333", f.InputDigit('3'))
	require.Equal(t, "333 3", f.InputDigit('3'))
	require.Equal(t, "333 33", f.InputDigit('3'))
	require.Equal(t, "333 333", f.InputDigit('3'))
}

func TestAYTFShortNumberFormattingFix_US(t *testing.T) {
	useTestMetadata(t)
	// For the US, an initial 1 is treated specially.
	f := GetAsYouTypeFormatter(regionCode.US)

	// 101 - Test that the initial 1 is not treated as a national prefix.
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "10", f.InputDigit('0'))
	require.Equal(t, "101", f.InputDigit('1'))

	// 112 - Test that the initial 1 is not treated as a national prefix.
	f.Clear()
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "11", f.InputDigit('1'))
	require.Equal(t, "112", f.InputDigit('2'))

	// 122 - Test that the initial 1 is treated as a national prefix.
	f.Clear()
	require.Equal(t, "1", f.InputDigit('1'))
	require.Equal(t, "12", f.InputDigit('2'))
	require.Equal(t, "1 22", f.InputDigit('2'))
}

func TestAYTFClearNDDAfterIDDExtraction(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.KR)

	// Check that when we have successfully extracted an IDD, the previously extracted
	// NDD is cleared since it is no longer valid.
	require.Equal(t, "0", f.InputDigit('0'))
	require.Equal(t, "00", f.InputDigit('0'))
	require.Equal(t, "007", f.InputDigit('7'))
	require.Equal(t, "0070", f.InputDigit('0'))
	require.Equal(t, "00700", f.InputDigit('0'))
	require.Equal(t, "0", f.getExtractedNationalPrefix())

	// Once the IDD "00700" has been extracted, it no longer makes sense for the initial
	// "0" to be treated as an NDD.
	require.Equal(t, "00700 1 ", f.InputDigit('1'))
	require.Equal(t, "", f.getExtractedNationalPrefix())

	require.Equal(t, "00700 1 2", f.InputDigit('2'))
	require.Equal(t, "00700 1 23", f.InputDigit('3'))
	require.Equal(t, "00700 1 234", f.InputDigit('4'))
	require.Equal(t, "00700 1 234 5", f.InputDigit('5'))
	require.Equal(t, "00700 1 234 56", f.InputDigit('6'))
	require.Equal(t, "00700 1 234 567", f.InputDigit('7'))
	require.Equal(t, "00700 1 234 567 8", f.InputDigit('8'))
	require.Equal(t, "00700 1 234 567 89", f.InputDigit('9'))
	require.Equal(t, "00700 1 234 567 890", f.InputDigit('0'))
	require.Equal(t, "00700 1 234 567 8901", f.InputDigit('1'))
	require.Equal(t, "00700123456789012", f.InputDigit('2'))
	require.Equal(t, "007001234567890123", f.InputDigit('3'))
	require.Equal(t, "0070012345678901234", f.InputDigit('4'))
	require.Equal(t, "00700123456789012345", f.InputDigit('5'))
	require.Equal(t, "007001234567890123456", f.InputDigit('6'))
	require.Equal(t, "0070012345678901234567", f.InputDigit('7'))
}

func TestAYTFNumberPatternsBecomingInvalidShouldNotResultInDigitLoss(t *testing.T) {
	useTestMetadata(t)
	f := GetAsYouTypeFormatter(regionCode.CN)

	require.Equal(t, "+", f.InputDigit('+'))
	require.Equal(t, "+8", f.InputDigit('8'))
	require.Equal(t, "+86 ", f.InputDigit('6'))
	require.Equal(t, "+86 9", f.InputDigit('9'))
	require.Equal(t, "+86 98", f.InputDigit('8'))
	require.Equal(t, "+86 988", f.InputDigit('8'))
	require.Equal(t, "+86 988 1", f.InputDigit('1'))
	// Now the number pattern is no longer valid because there are multiple leading
	// digit patterns; when we try again to extract a country code we should ensure we
	// use the last leading digit pattern, rather than the first one such that it
	// *thinks* it's found a valid formatting rule again.
	// https://github.com/google/libphonenumber/issues/437
	require.Equal(t, "+8698812", f.InputDigit('2'))
	require.Equal(t, "+86988123", f.InputDigit('3'))
	require.Equal(t, "+869881234", f.InputDigit('4'))
	require.Equal(t, "+8698812345", f.InputDigit('5'))
}
