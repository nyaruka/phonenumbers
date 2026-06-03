package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java
// possibility / truncation methods, run against the synthetic test metadata.
// Method names and assertions mirror the Java.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

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
