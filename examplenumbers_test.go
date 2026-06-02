package phonenumbers

// Faithful port of upstream libphonenumber's ExampleNumbersTest.java.
//
// Unlike the PhoneNumberUtilTest ports, ExampleNumbersTest is a real-metadata
// regression: upstream uses PhoneNumberUtil.getInstance() / ShortNumberInfo.
// getInstance() (production metadata), so these run against the embedded metadata
// and assert that every shipped region's/type's example numbers parse, validate,
// and classify correctly. Do not call useTestMetadata here.
//
// Test names mirror the upstream method names. The one exception is
// TestCanBeInternationallyDialledExampleNumbers: upstream has a
// testCanBeInternationallyDialled in both ExampleNumbersTest and
// PhoneNumberUtilTest, and the latter is already ported as
// TestCanBeInternationallyDialled in phonenumberutil_test.go, so this variant
// carries a suffix to disambiguate within the flat Go package.

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkNumbersValidAndCorrectType requests an example number of exampleNumberRequestedType for
// every supported region and returns those that fail validation, and those whose detected type is
// not in possibleExpectedTypes.
func checkNumbersValidAndCorrectType(exampleNumberRequestedType PhoneNumberType, possibleExpectedTypes map[PhoneNumberType]bool) (invalidCases, wrongTypeCases []*PhoneNumber) {
	for regionCode := range GetSupportedRegions() {
		exampleNumber := GetExampleNumberForType(regionCode, exampleNumberRequestedType)
		if exampleNumber == nil {
			continue
		}
		if !IsValidNumber(exampleNumber) {
			invalidCases = append(invalidCases, exampleNumber)
		} else if !possibleExpectedTypes[GetNumberType(exampleNumber)] {
			// We know the number is valid, now we check the type.
			wrongTypeCases = append(wrongTypeCases, exampleNumber)
		}
	}
	return invalidCases, wrongTypeCases
}

func TestFixedLine(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(FIXED_LINE, map[PhoneNumberType]bool{FIXED_LINE: true, FIXED_LINE_OR_MOBILE: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestMobile(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(MOBILE, map[PhoneNumberType]bool{MOBILE: true, FIXED_LINE_OR_MOBILE: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestTollFree(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(TOLL_FREE, map[PhoneNumberType]bool{TOLL_FREE: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestPremiumRate(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(PREMIUM_RATE, map[PhoneNumberType]bool{PREMIUM_RATE: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestVoip(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(VOIP, map[PhoneNumberType]bool{VOIP: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestPager(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(PAGER, map[PhoneNumberType]bool{PAGER: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestUan(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(UAN, map[PhoneNumberType]bool{UAN: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestVoicemail(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(VOICEMAIL, map[PhoneNumberType]bool{VOICEMAIL: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestSharedCost(t *testing.T) {
	invalid, wrongType := checkNumbersValidAndCorrectType(SHARED_COST, map[PhoneNumberType]bool{SHARED_COST: true})
	assert.Empty(t, invalid)
	assert.Empty(t, wrongType)
}

func TestCanBeInternationallyDialledExampleNumbers(t *testing.T) {
	var wrongTypeCases []*PhoneNumber
	for regionCode := range GetSupportedRegions() {
		desc := getMetadataForRegion(regionCode).GetNoInternationalDialling()
		if desc.GetExampleNumber() == "" {
			continue
		}
		exampleNumber, err := Parse(desc.GetExampleNumber(), regionCode)
		if err != nil {
			t.Errorf("error parsing no-international-dialling example for %s: %s", regionCode, err)
			continue
		}
		if canBeInternationallyDialled(exampleNumber) {
			wrongTypeCases = append(wrongTypeCases, exampleNumber)
		}
	}
	assert.Empty(t, wrongTypeCases)
}

func TestGlobalNetworkNumbers(t *testing.T) {
	var invalidCases []*PhoneNumber
	for callingCode := range GetSupportedGlobalNetworkCallingCodes() {
		exampleNumber := GetExampleNumberForNonGeoEntity(callingCode)
		if !assert.NotNilf(t, exampleNumber, "No example phone number for calling code %d", callingCode) {
			continue
		}
		if !IsValidNumber(exampleNumber) {
			invalidCases = append(invalidCases, exampleNumber)
		}
	}
	assert.Empty(t, invalidCases)
}

func TestEveryRegionHasAnExampleNumber(t *testing.T) {
	for regionCode := range GetSupportedRegions() {
		assert.NotNilf(t, GetExampleNumber(regionCode), "No example number found for region %s", regionCode)
	}
}

func TestEveryRegionHasAnInvalidExampleNumber(t *testing.T) {
	for regionCode := range GetSupportedRegions() {
		assert.NotNilf(t, GetInvalidExampleNumber(regionCode), "No invalid example number found for region %s", regionCode)
	}
}

func TestEveryTypeHasAnExampleNumber(t *testing.T) {
	for _, typ := range allPhoneNumberTypes {
		if typ == UNKNOWN {
			continue
		}
		assert.NotNilf(t, getExampleNumberForTypeAnyRegion(t, typ), "No example number found for type %v", typ)
	}
}

// getExampleNumberForTypeAnyRegion is the region-less analogue of upstream's
// getExampleNumberForType(type) (which the Go package does not expose): it returns
// an example number of the given type from the first supported region that has
// one, falling back to non-geographical entities.
func getExampleNumberForTypeAnyRegion(t *testing.T, typ PhoneNumberType) *PhoneNumber {
	for regionCode := range GetSupportedRegions() {
		if exampleNumber := GetExampleNumberForType(regionCode, typ); exampleNumber != nil {
			return exampleNumber
		}
	}
	for callingCode := range GetSupportedGlobalNetworkCallingCodes() {
		desc := getNumberDescByType(getMetadataForNonGeographicalRegion(callingCode), typ)
		if desc.GetExampleNumber() == "" {
			continue
		}
		exampleNumber, err := Parse("+"+strconv.Itoa(callingCode)+desc.GetExampleNumber(), UNKNOWN_REGION)
		if err != nil {
			t.Logf("error parsing non-geo example for calling code %d: %s", callingCode, err)
			continue
		}
		return exampleNumber
	}
	return nil
}

func TestShortNumbersValidAndCorrectCost(t *testing.T) {
	var invalidStringCases []string
	var invalidCases, wrongTypeCases []*PhoneNumber
	for regionCode := range shortNumberRegionToMetadataMap {
		exampleShortNumber := getExampleShortNumber(regionCode)
		number, err := Parse(exampleShortNumber, regionCode)
		require.NoErrorf(t, err, "parsing example short number %q for %s", exampleShortNumber, regionCode)
		if !IsValidShortNumberForRegion(number, regionCode) {
			invalidStringCases = append(invalidStringCases, "region_code: "+regionCode+", national_number: "+exampleShortNumber)
		}
		if !IsValidShortNumber(number) {
			invalidCases = append(invalidCases, number)
		}

		for _, cost := range []ShortNumberCost{TOLL_FREE_COST, STANDARD_RATE_COST, PREMIUM_RATE_COST, UNKNOWN_COST} {
			exampleShortNumber = getExampleShortNumberForCost(regionCode, cost)
			if exampleShortNumber == "" {
				continue
			}
			number, err = Parse(exampleShortNumber, regionCode)
			require.NoErrorf(t, err, "parsing example short number %q for %s", exampleShortNumber, regionCode)
			if cost != GetExpectedCostForRegion(number, regionCode) {
				wrongTypeCases = append(wrongTypeCases, number)
			}
		}
	}
	assert.Empty(t, invalidStringCases)
	assert.Empty(t, invalidCases)
	assert.Empty(t, wrongTypeCases)
}

func TestEmergency(t *testing.T) {
	wrongTypeCounter := 0
	for regionCode := range shortNumberRegionToMetadataMap {
		desc := getShortNumberMetadataForRegion(regionCode).GetEmergency()
		if desc.GetExampleNumber() == "" {
			continue
		}
		exampleNumber := desc.GetExampleNumber()
		number, err := Parse(exampleNumber, regionCode)
		require.NoErrorf(t, err, "parsing emergency example %q for %s", exampleNumber, regionCode)
		if !IsPossibleShortNumberForRegion(number, regionCode) || !IsEmergencyNumber(exampleNumber, regionCode) {
			wrongTypeCounter++
		} else if GetExpectedCostForRegion(number, regionCode) != TOLL_FREE_COST {
			wrongTypeCounter++
		}
	}
	assert.Equal(t, 0, wrongTypeCounter)
}

func TestCarrierSpecificShortNumbers(t *testing.T) {
	wrongTagCounter := 0
	for regionCode := range shortNumberRegionToMetadataMap {
		desc := getShortNumberMetadataForRegion(regionCode).GetCarrierSpecific()
		if desc.GetExampleNumber() == "" {
			continue
		}
		exampleNumber := desc.GetExampleNumber()
		carrierSpecificNumber, err := Parse(exampleNumber, regionCode)
		require.NoErrorf(t, err, "parsing carrier-specific example %q for %s", exampleNumber, regionCode)
		if !IsPossibleShortNumberForRegion(carrierSpecificNumber, regionCode) || !IsCarrierSpecificForRegion(carrierSpecificNumber, regionCode) {
			wrongTagCounter++
		}
	}
	assert.Equal(t, 0, wrongTagCounter)
}

func TestSmsServiceShortNumbers(t *testing.T) {
	wrongTagCounter := 0
	for regionCode := range shortNumberRegionToMetadataMap {
		desc := getShortNumberMetadataForRegion(regionCode).GetSmsServices()
		if desc.GetExampleNumber() == "" {
			continue
		}
		exampleNumber := desc.GetExampleNumber()
		smsServiceNumber, err := Parse(exampleNumber, regionCode)
		require.NoErrorf(t, err, "parsing sms-service example %q for %s", exampleNumber, regionCode)
		if !IsPossibleShortNumberForRegion(smsServiceNumber, regionCode) || !IsSmsServiceForRegion(smsServiceNumber, regionCode) {
			wrongTagCounter++
		}
	}
	assert.Equal(t, 0, wrongTagCounter)
}
