package phonenumbers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

////////// Copied from java-libphonenumber
/**
 * Unit tests for ShortNumberInfo.java
 */

func TestIsPossibleShortNumber(t *testing.T) {
	countryCode := int32(33)
	nationalNumber := uint64(123456)
	possibleNumber := &PhoneNumber{
		CountryCode:    &countryCode,
		NationalNumber: &nationalNumber,
	}
	assert.True(t, IsPossibleShortNumber(possibleNumber))

	possibleNumber, err := Parse("123456", "FR")
	if err != nil {
		t.Errorf("Error parsing number: %s: %s", "123456", err)
	}
	assert.True(t, IsPossibleShortNumberForRegion(possibleNumber, "FR"))

	nationalNumber = 9
	impossibleNumber := &PhoneNumber{
		CountryCode:    &countryCode,
		NationalNumber: &nationalNumber,
	}
	assert.False(t, IsPossibleShortNumber(impossibleNumber))

	// Note that GB and GG share the country calling code 44, and that this number is possible but
	// not valid.
	countryCode = 44
	nationalNumber = 11001
	possibleNumber = &PhoneNumber{
		CountryCode:    &countryCode,
		NationalNumber: &nationalNumber,
	}
	assert.True(t, IsPossibleShortNumber(possibleNumber))
}

func TestIsValidShortNumber(t *testing.T) {
	countryCode := int32(33)
	nationalNumber := uint64(1010)
	validNumber := &PhoneNumber{
		CountryCode:    &countryCode,
		NationalNumber: &nationalNumber,
	}
	assert.True(t, IsValidShortNumber(validNumber))

	validNumber, err := Parse("1010", "FR")
	if err != nil {
		t.Errorf("Error parsing number: %s: %s", "1010", err)
	}
	assert.True(t, IsValidShortNumberForRegion(validNumber, "FR"))

	nationalNumber = uint64(123456)
	invalidNumber := &PhoneNumber{
		CountryCode:    &countryCode,
		NationalNumber: &nationalNumber,
	}
	assert.False(t, IsValidShortNumber(invalidNumber))

	invalidNumber, err = Parse("123456", "FR")
	if err != nil {
		t.Errorf("Error parsing number: %s: %s", "1010", err)
	}
	assert.False(t, IsValidShortNumberForRegion(invalidNumber, "FR"))
}

func TestConnectsToEmergencyNumber_US(t *testing.T) {
	assert.True(t, ConnectsToEmergencyNumber("911", "US"))
	assert.True(t, ConnectsToEmergencyNumber("112", "US"))
	assert.False(t, ConnectsToEmergencyNumber("999", "US"))
}

func TestConnectsToEmergencyLongNumber_US(t *testing.T) {
	assert.True(t, ConnectsToEmergencyNumber("9116666666", "US"))
	assert.True(t, ConnectsToEmergencyNumber("1126666666", "US"))
	assert.False(t, ConnectsToEmergencyNumber("9996666666", "US"))
}

func TestConnectsToEmergencyNumberWithFormatting_US(t *testing.T) {

	assert.True(t, ConnectsToEmergencyNumber("9-1-1", "US"))
	assert.True(t, ConnectsToEmergencyNumber("1-1-2", "US"))
	assert.False(t, ConnectsToEmergencyNumber("9-9-9", "US"))
}

func TestConnectsToEmergencyNumber_BR(t *testing.T) {
	assert.True(t, ConnectsToEmergencyNumber("190", "BR"))
	assert.True(t, ConnectsToEmergencyNumber("911", "BR"))
	assert.False(t, ConnectsToEmergencyNumber("999", "BR"))
}

func TestConnectsToEmergencyNumberLongNumber_BR(t *testing.T) {
	assert.False(t, ConnectsToEmergencyNumber("9111", "BR"))
	assert.False(t, ConnectsToEmergencyNumber("1900", "BR"))
	assert.False(t, ConnectsToEmergencyNumber("9996", "BR"))
}

func TestConnectsToEmergencyNumber_CL(t *testing.T) {
	assert.True(t, ConnectsToEmergencyNumber("131", "CL"))
	assert.True(t, ConnectsToEmergencyNumber("133", "CL"))
}

func TestConnectsToEmergencyNumberLongNumber_CL(t *testing.T) {
	assert.False(t, ConnectsToEmergencyNumber("1313", "CL"))
	assert.False(t, ConnectsToEmergencyNumber("1330", "CL"))
}

func TestConnectsToEmergencyNumber_AO(t *testing.T) {
	assert.False(t, ConnectsToEmergencyNumber("911", "AO"))
	assert.False(t, ConnectsToEmergencyNumber("222123456", "AO"))
	assert.False(t, ConnectsToEmergencyNumber("923123456", "AO"))
}

func TestConnectsToEmergencyNumber_ZW(t *testing.T) {
	assert.False(t, ConnectsToEmergencyNumber("911", "ZW"))
	assert.False(t, ConnectsToEmergencyNumber("01312345", "ZW"))
	assert.False(t, ConnectsToEmergencyNumber("0711234567", "ZW"))
}

func TestIsEmergencyNumber_US(t *testing.T) {
	assert.True(t, IsEmergencyNumber("911", "US"))
	assert.True(t, IsEmergencyNumber("112", "US"))
	assert.False(t, IsEmergencyNumber("999", "US"))
}

func TestIsEmergencyNumberLongNumber_US(t *testing.T) {
	assert.False(t, IsEmergencyNumber("9116666666", "US"))
	assert.False(t, IsEmergencyNumber("1126666666", "US"))
	assert.False(t, IsEmergencyNumber("9996666666", "US"))
}

func TestIsEmergencyNumberWithFormatting_US(t *testing.T) {
	assert.True(t, IsEmergencyNumber("9-1-1", "US"))
	assert.True(t, IsEmergencyNumber("*911", "US"))
	assert.True(t, IsEmergencyNumber("1-1-2", "US"))
	assert.True(t, IsEmergencyNumber("*112", "US"))
	assert.False(t, IsEmergencyNumber("9-9-9", "US"))
	assert.False(t, IsEmergencyNumber("*999", "US"))
}

func TestIsEmergencyNumberWithPlusSign_US(t *testing.T) {
	assert.False(t, IsEmergencyNumber("+911", "US"))
	assert.False(t, IsEmergencyNumber("\uFF0B911", "US"))
	assert.False(t, IsEmergencyNumber(" +911", "US"))
	assert.False(t, IsEmergencyNumber("+112", "US"))
	assert.False(t, IsEmergencyNumber("+999", "US"))
}

func TestIsEmergencyNumber_BR(t *testing.T) {
	assert.True(t, IsEmergencyNumber("911", "BR"))
	assert.True(t, IsEmergencyNumber("190", "BR"))
	assert.False(t, IsEmergencyNumber("999", "BR"))
}

func TestIsEmergencyNumberLongNumber_BR(t *testing.T) {
	assert.False(t, IsEmergencyNumber("9111", "BR"))
	assert.False(t, IsEmergencyNumber("1900", "BR"))
	assert.False(t, IsEmergencyNumber("9996", "BR"))
}

func TestIsEmergencyNumber_AO(t *testing.T) {
	assert.False(t, IsEmergencyNumber("911", "AO"))
	assert.False(t, IsEmergencyNumber("222123456", "AO"))
	assert.False(t, IsEmergencyNumber("923123456", "AO"))
}

func TestIsEmergencyNumber_ZW(t *testing.T) {
	assert.False(t, IsEmergencyNumber("911", "ZW"))
	assert.False(t, IsEmergencyNumber("01312345", "ZW"))
	assert.False(t, IsEmergencyNumber("0711234567", "ZW"))
}

func TestEmergencyNumberForSharedCountryCallingCode(t *testing.T) {
	assert.True(t, IsEmergencyNumber("112", "AU"))
	assert.True(t, IsValidShortNumberForRegion(parse(t, "112", "AU"), "AU"))
	assert.True(t, IsEmergencyNumber("112", "CX"))
	assert.True(t, IsValidShortNumberForRegion(parse(t, "112", "CX"), "CX"))
	sharedEmergencyNumber := &PhoneNumber{
		CountryCode:    proto.Int32(61),
		NationalNumber: proto.Uint64(112),
	}
	assert.True(t, IsValidShortNumber(sharedEmergencyNumber))
}

func TestOverlappingNANPANumber(t *testing.T) {
	assert.True(t, IsEmergencyNumber("211", "BB"))
	assert.False(t, IsEmergencyNumber("211", "US"))
	assert.False(t, IsEmergencyNumber("211", "CA"))
}

func TestCountryCallingCodeIsNotIgnored(t *testing.T) {
	assert.False(t, IsPossibleShortNumberForRegion(parse(t, "+4640404", "SE"), "US"))
	assert.False(t, IsValidShortNumberForRegion(parse(t, "+4640404", "SE"), "US"))
}

func parse(t *testing.T, number string, regionCode string) *PhoneNumber {
	phoneNumber, err := Parse(number, regionCode)
	if err != nil {
		t.Fatalf("Test input data should always parse correctly: %s (%s)", number, regionCode)
	}
	return phoneNumber
}
