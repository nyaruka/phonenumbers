package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java
// normalization / alpha / viability / extraction methods. These exercise static
// helpers that don't touch metadata, so (unlike most ports) they don't call
// useTestMetadata. Method names and assertions mirror the Java; non-ASCII
// inputs use the same \u escapes as upstream.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
