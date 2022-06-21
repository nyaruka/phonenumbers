package phonenumbers

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
