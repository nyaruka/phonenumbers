package shortnumber

import (
	"fmt"
	"testing"

	"github.com/nyaruka/phonenumbers"
)

func TestFoo(t *testing.T) {
	if !isValidRegionCode("SE") {
		t.FailNow()
	}
}

func TestIsEmergencyNumber(t *testing.T) {
	shouldMatch := func(number string, regionCode string) {
		t.Run(fmt.Sprintf("IsEmergencyNumber(%s, %s)", number, regionCode), func(t *testing.T) {
			if !IsEmergencyNumber(number, regionCode) {
				t.Fatalf("%q is not a valid emergency number in %q", number, regionCode)
			}
			if !ConnectsToEmergencyNumber(number, regionCode) {
				t.Fatalf("%q does not connect to the emergency number despite it being a valid emergency number in %q", number, regionCode)
			}
		})
	}
	shouldNotMatch := func(number string, regionCode string) {
		t.Run(fmt.Sprintf("!IsEmergencyNumber(%s, %s)", number, regionCode), func(t *testing.T) {
			if IsEmergencyNumber(number, regionCode) {
				t.Fatalf("%q should not be a valid emergency number in %q", number, regionCode)
			}
		})
	}

	shouldMatch("911", "US")
	shouldMatch("112", "US")
	shouldMatch("112", "SE")
	shouldMatch("90000", "SE")
	// shouldNotMatch("+1911", "US") // wut?

	shouldNotMatch("911", "SE")
	shouldNotMatch("9111", "US")
	shouldNotMatch("12345", "SE")
}

func TestConnectsToEmergencyNumber(t *testing.T) {
	shouldMatch := func(number string, regionCode string) {
		t.Run(fmt.Sprintf("ConnectsToEmergencyNumber(%s, %s)", number, regionCode), func(t *testing.T) {
			if !ConnectsToEmergencyNumber(number, regionCode) {
				t.Fatalf("%q should connect to the emergency number in %q", number, regionCode)
			}
		})
	}
	shouldNotMatch := func(number string, regionCode string) {
		t.Run(fmt.Sprintf("!ConnectsToEmergencyNumber(%s, %s)", number, regionCode), func(t *testing.T) {
			if ConnectsToEmergencyNumber(number, regionCode) {
				t.Fatalf("%q should not connect to the emergency number in %q", number, regionCode)
			}
		})
	}

	shouldMatch("911", "US")
	shouldMatch("112", "US")
	shouldMatch("112", "SE")
	shouldMatch("90000", "SE")
	shouldMatch("9116666666", "US")
	shouldMatch("1126666666", "US")

	shouldNotMatch("9996666666", "US")

	shouldNotMatch("911", "SE")
	shouldNotMatch("12345", "SE")

	// Brazilian emergency numbers don't work when additional digits are appended.
	shouldNotMatch("9111", "BR")
	shouldNotMatch("1900", "BR")
	shouldNotMatch("9996", "BR")
}

func TestGetExpectedCost(t *testing.T) {
	testFRNumber := func(t *testing.T, number string, rate ShortNumberCost) {
		parsedNumber, err := phonenumbers.Parse(number, "FR")
		if err != nil {
			t.Fatalf("failed to parse example phone number: %+v", err)
		}
		result := GetExpectedCostForRegion(parsedNumber, "FR")
		if result != rate {
			t.Fatalf("Got cost %v while expecting %v", result, rate)
		}
		// ... something something shortInfo.getExpectedCost(premiumRateNumber));
	}

	testFRRateExample := func(t *testing.T, rate ShortNumberCost) {
		rateExample := getExampleShortNumberForCost("FR", rate)
		testFRNumber(t, rateExample, rate)
	}
	t.Run("FR PREMIUM_RATE", func(t *testing.T) { testFRRateExample(t, PREMIUM_RATE) })
	t.Run("FR STANDARD_RATE", func(t *testing.T) { testFRRateExample(t, STANDARD_RATE) })
	t.Run("FR TOLL_FREE", func(t *testing.T) { testFRRateExample(t, TOLL_FREE) })

	t.Run("FR UNKNOWN_COST", func(t *testing.T) { testFRNumber(t, "12345", UNKNOWN_COST) })
	t.Run("FR Emergency", func(t *testing.T) { testFRNumber(t, "112", TOLL_FREE) })
}

func getExampleShortNumberForCost(regionCode string, cost ShortNumberCost) string {
	phoneMetadata := getMetadataForRegion(regionCode)
	if phoneMetadata == nil {
		return ""
	}

	var desc *phonenumbers.PhoneNumberDesc
	switch cost {
	case TOLL_FREE:
		desc = phoneMetadata.GetTollFree()
	case STANDARD_RATE:
		desc = phoneMetadata.GetStandardRate()
	case PREMIUM_RATE:
		desc = phoneMetadata.GetPremiumRate()
	default:
		// UNKNOWN_COST numbers are computed by the process of elimination from the other cost
		// categories.
	}
	return desc.GetExampleNumber()
}

/*
  String getExampleShortNumberForCost(String regionCode, ShortNumberCost cost) {
    PhoneMetadata phoneMetadata = MetadataManager.getShortNumberMetadataForRegion(regionCode);
    if (phoneMetadata == null) {
      return "";
    }
    PhoneNumberDesc desc = null;
    switch (cost) {
      case TOLL_FREE:
        desc = phoneMetadata.getTollFree();
        break;
      case STANDARD_RATE:
        desc = phoneMetadata.getStandardRate();
        break;
      case PREMIUM_RATE:
        desc = phoneMetadata.getPremiumRate();
        break;
      default:
        // UNKNOWN_COST numbers are computed by the process of elimination from the other cost
        // categories.
    }
    if (desc != null && desc.hasExampleNumber()) {
      return desc.getExampleNumber();
    }
    return "";
  }
*/
