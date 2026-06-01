package phonenumbers

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// Faithful port of upstream libphonenumber's ShortNumberInfoTest.java.
// Unlike PhoneNumberUtilTest, upstream's ShortNumberInfoTest has no synthetic
// metadata file: ShortNumberInfo.getInstance() always uses the production short
// metadata, so these run against the embedded short-number metadata (just like
// upstream). They are therefore real-metadata regressions, not synthetic.
// Last reconciled against: v9.0.32

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

func TestIsCarrierSpecific(t *testing.T) {
	carrierSpecificNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(33669)}
	assert.True(t, IsCarrierSpecific(carrierSpecificNumber))
	assert.True(t, IsCarrierSpecificForRegion(parse(t, "33669", "US"), "US"))

	notCarrierSpecificNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(911)}
	assert.False(t, IsCarrierSpecific(notCarrierSpecificNumber))
	assert.False(t, IsCarrierSpecificForRegion(parse(t, "911", "US"), "US"))

	carrierSpecificNumberForSomeRegion := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(211)}
	assert.True(t, IsCarrierSpecific(carrierSpecificNumberForSomeRegion))
	assert.True(t, IsCarrierSpecificForRegion(carrierSpecificNumberForSomeRegion, "US"))
	assert.False(t, IsCarrierSpecificForRegion(carrierSpecificNumberForSomeRegion, "BB"))
}

func TestIsSmsService(t *testing.T) {
	// SKIP: the committed embedded short metadata carries no smsServices data — it
	// was generated before the builder learned to read the <smsServices> element
	// (see builder.go and TestBuilderProcessesSmsServices). IsSmsServiceForRegion
	// therefore returns false for every region against the embedded snapshot. This
	// un-skips once the short metadata is regenerated (a separate data refresh).
	t.Skip("blocked on short metadata predating smsServices builder support (needs data regen)")
	smsServiceNumberForSomeRegion := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(21234)}
	assert.True(t, IsSmsServiceForRegion(smsServiceNumberForSomeRegion, "US"))
	assert.False(t, IsSmsServiceForRegion(smsServiceNumberForSomeRegion, "BB"))
}

func TestGetExpectedCost(t *testing.T) {
	premiumRateExample := getExampleShortNumberForCost("FR", PREMIUM_RATE_COST)
	assert.Equal(t, PREMIUM_RATE_COST, GetExpectedCostForRegion(parse(t, premiumRateExample, "FR"), "FR"))
	premiumRateNumber := &PhoneNumber{CountryCode: proto.Int32(33), NationalNumber: proto.Uint64(atoui(t, premiumRateExample))}
	assert.Equal(t, PREMIUM_RATE_COST, GetExpectedCost(premiumRateNumber))

	standardRateExample := getExampleShortNumberForCost("FR", STANDARD_RATE_COST)
	assert.Equal(t, STANDARD_RATE_COST, GetExpectedCostForRegion(parse(t, standardRateExample, "FR"), "FR"))
	standardRateNumber := &PhoneNumber{CountryCode: proto.Int32(33), NationalNumber: proto.Uint64(atoui(t, standardRateExample))}
	assert.Equal(t, STANDARD_RATE_COST, GetExpectedCost(standardRateNumber))

	tollFreeExample := getExampleShortNumberForCost("FR", TOLL_FREE_COST)
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCostForRegion(parse(t, tollFreeExample, "FR"), "FR"))
	tollFreeNumber := &PhoneNumber{CountryCode: proto.Int32(33), NationalNumber: proto.Uint64(atoui(t, tollFreeExample))}
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCost(tollFreeNumber))

	assert.Equal(t, UNKNOWN_COST, GetExpectedCostForRegion(parse(t, "12345", "FR"), "FR"))
	unknownCostNumber := &PhoneNumber{CountryCode: proto.Int32(33), NationalNumber: proto.Uint64(12345)}
	assert.Equal(t, UNKNOWN_COST, GetExpectedCost(unknownCostNumber))

	// Test that an invalid number may nevertheless have a cost other than UNKNOWN_COST.
	assert.False(t, IsValidShortNumberForRegion(parse(t, "116123", "FR"), "FR"))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCostForRegion(parse(t, "116123", "FR"), "FR"))
	invalidNumber := &PhoneNumber{CountryCode: proto.Int32(33), NationalNumber: proto.Uint64(116123)}
	assert.False(t, IsValidShortNumber(invalidNumber))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCost(invalidNumber))

	// Test a nonexistent country code.
	assert.Equal(t, UNKNOWN_COST, GetExpectedCostForRegion(parse(t, "911", "US"), "ZZ"))
	unknownCostNumber = &PhoneNumber{CountryCode: proto.Int32(123), NationalNumber: proto.Uint64(911)}
	assert.Equal(t, UNKNOWN_COST, GetExpectedCost(unknownCostNumber))
}

func TestGetExpectedCostForSharedCountryCallingCode(t *testing.T) {
	// Test some numbers which have different costs in countries sharing the same country calling
	// code. In Australia, 1234 is premium-rate, 1194 is standard-rate, and 733 is toll-free. These
	// are not known to be valid numbers in the Christmas Islands.
	ambiguousPremiumRateString := "1234"
	ambiguousPremiumRateNumber := &PhoneNumber{CountryCode: proto.Int32(61), NationalNumber: proto.Uint64(1234)}
	ambiguousStandardRateString := "1194"
	ambiguousStandardRateNumber := &PhoneNumber{CountryCode: proto.Int32(61), NationalNumber: proto.Uint64(1194)}
	ambiguousTollFreeString := "733"
	ambiguousTollFreeNumber := &PhoneNumber{CountryCode: proto.Int32(61), NationalNumber: proto.Uint64(733)}

	assert.True(t, IsValidShortNumber(ambiguousPremiumRateNumber))
	assert.True(t, IsValidShortNumber(ambiguousStandardRateNumber))
	assert.True(t, IsValidShortNumber(ambiguousTollFreeNumber))

	assert.True(t, IsValidShortNumberForRegion(parse(t, ambiguousPremiumRateString, "AU"), "AU"))
	assert.Equal(t, PREMIUM_RATE_COST, GetExpectedCostForRegion(parse(t, ambiguousPremiumRateString, "AU"), "AU"))
	assert.False(t, IsValidShortNumberForRegion(parse(t, ambiguousPremiumRateString, "CX"), "CX"))
	assert.Equal(t, UNKNOWN_COST, GetExpectedCostForRegion(parse(t, ambiguousPremiumRateString, "CX"), "CX"))
	// PREMIUM_RATE takes precedence over UNKNOWN_COST.
	assert.Equal(t, PREMIUM_RATE_COST, GetExpectedCost(ambiguousPremiumRateNumber))

	assert.True(t, IsValidShortNumberForRegion(parse(t, ambiguousStandardRateString, "AU"), "AU"))
	assert.Equal(t, STANDARD_RATE_COST, GetExpectedCostForRegion(parse(t, ambiguousStandardRateString, "AU"), "AU"))
	assert.False(t, IsValidShortNumberForRegion(parse(t, ambiguousStandardRateString, "CX"), "CX"))
	assert.Equal(t, UNKNOWN_COST, GetExpectedCostForRegion(parse(t, ambiguousStandardRateString, "CX"), "CX"))
	assert.Equal(t, UNKNOWN_COST, GetExpectedCost(ambiguousStandardRateNumber))

	assert.True(t, IsValidShortNumberForRegion(parse(t, ambiguousTollFreeString, "AU"), "AU"))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCostForRegion(parse(t, ambiguousTollFreeString, "AU"), "AU"))
	assert.False(t, IsValidShortNumberForRegion(parse(t, ambiguousTollFreeString, "CX"), "CX"))
	assert.Equal(t, UNKNOWN_COST, GetExpectedCostForRegion(parse(t, ambiguousTollFreeString, "CX"), "CX"))
	assert.Equal(t, UNKNOWN_COST, GetExpectedCost(ambiguousTollFreeNumber))
}

func TestExampleShortNumberPresence(t *testing.T) {
	assert.NotEmpty(t, getExampleShortNumber("AD"))
	assert.NotEmpty(t, getExampleShortNumber("FR"))
	assert.Empty(t, getExampleShortNumber("001"))
	assert.Empty(t, getExampleShortNumber(""))
}

func TestConnectsToEmergencyNumber_US(t *testing.T) {
	assert.True(t, ConnectsToEmergencyNumber("911", "US"))
	assert.True(t, ConnectsToEmergencyNumber("112", "US"))
	assert.False(t, ConnectsToEmergencyNumber("999", "US"))
}

func TestConnectsToEmergencyNumberLongNumber_US(t *testing.T) {
	assert.True(t, ConnectsToEmergencyNumber("9116666666", "US"))
	assert.True(t, ConnectsToEmergencyNumber("1126666666", "US"))
	assert.False(t, ConnectsToEmergencyNumber("9996666666", "US"))
}

func TestConnectsToEmergencyNumberWithFormatting_US(t *testing.T) {
	assert.True(t, ConnectsToEmergencyNumber("9-1-1", "US"))
	assert.True(t, ConnectsToEmergencyNumber("1-1-2", "US"))
	assert.False(t, ConnectsToEmergencyNumber("9-9-9", "US"))
}

func TestConnectsToEmergencyNumberWithPlusSign_US(t *testing.T) {
	assert.False(t, ConnectsToEmergencyNumber("+911", "US"))
	assert.False(t, ConnectsToEmergencyNumber("＋911", "US"))
	assert.False(t, ConnectsToEmergencyNumber(" +911", "US"))
	assert.False(t, ConnectsToEmergencyNumber("+112", "US"))
	assert.False(t, ConnectsToEmergencyNumber("+999", "US"))
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
	// Test the emergency number 112, which is valid in both Australia and the Christmas Islands.
	assert.True(t, IsEmergencyNumber("112", "AU"))
	assert.True(t, IsValidShortNumberForRegion(parse(t, "112", "AU"), "AU"))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCostForRegion(parse(t, "112", "AU"), "AU"))
	assert.True(t, IsEmergencyNumber("112", "CX"))
	assert.True(t, IsValidShortNumberForRegion(parse(t, "112", "CX"), "CX"))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCostForRegion(parse(t, "112", "CX"), "CX"))
	sharedEmergencyNumber := &PhoneNumber{
		CountryCode:    proto.Int32(61),
		NationalNumber: proto.Uint64(112),
	}
	assert.True(t, IsValidShortNumber(sharedEmergencyNumber))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCost(sharedEmergencyNumber))
}

func TestOverlappingNANPANumber(t *testing.T) {
	// 211 is an emergency number in Barbados, while it is a toll-free information line in Canada
	// and the USA.
	assert.True(t, IsEmergencyNumber("211", "BB"))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCostForRegion(parse(t, "211", "BB"), "BB"))
	assert.False(t, IsEmergencyNumber("211", "US"))
	assert.Equal(t, UNKNOWN_COST, GetExpectedCostForRegion(parse(t, "211", "US"), "US"))
	assert.False(t, IsEmergencyNumber("211", "CA"))
	assert.Equal(t, TOLL_FREE_COST, GetExpectedCostForRegion(parse(t, "211", "CA"), "CA"))
}

func TestCountryCallingCodeIsNotIgnored(t *testing.T) {
	// +46 is the country calling code for Sweden (SE), and 40404 is a valid short number in the US.
	assert.False(t, IsPossibleShortNumberForRegion(parse(t, "+4640404", "SE"), "US"))
	assert.False(t, IsValidShortNumberForRegion(parse(t, "+4640404", "SE"), "US"))
	assert.Equal(t, UNKNOWN_COST, GetExpectedCostForRegion(parse(t, "+4640404", "SE"), "US"))
}

// atoui parses a decimal string into a uint64, failing the test on error. Mirrors upstream's
// Integer.parseInt usage on example short numbers.
func atoui(t *testing.T, s string) uint64 {
	t.Helper()
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		t.Fatalf("expected a numeric short number, got %q: %s", s, err)
	}
	return n
}

func parse(t *testing.T, number string, regionCode string) *PhoneNumber {
	phoneNumber, err := Parse(number, regionCode)
	if err != nil {
		t.Fatalf("Test input data should always parse correctly: %s (%s)", number, regionCode)
	}
	return phoneNumber
}
