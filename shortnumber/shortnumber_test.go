package shortnumber

import (
	"fmt"
	"testing"
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
			if !IsEmergencyNumber(number, regionCode) {
				t.Fatalf("%q is a valid emergency number in %q", number, regionCode)
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
				t.Fatalf("%q does not connect to the emergency number in %q", number, regionCode)
			}
		})
	}
	shouldNotMatch := func(number string, regionCode string) {
		t.Run(fmt.Sprintf("!ConnectsToEmergencyNumber(%s, %s)", number, regionCode), func(t *testing.T) {
			if !ConnectsToEmergencyNumber(number, regionCode) {
				t.Fatalf("%q connects to the emergency number in %q", number, regionCode)
			}
		})
	}

	shouldMatch("911", "US")
	shouldMatch("112", "US")
	shouldMatch("112", "SE")
	shouldMatch("90000", "SE")
	shouldMatch("9116666666", "US")
	shouldMatch("1126666666", "US")

	shouldNotMatch("9111", "US")
	shouldNotMatch("9996666666", "US")

	shouldNotMatch("911", "SE")
	shouldNotMatch("12345", "SE")

	// Brazilian emergency numbers don't work when additional digits are appended.
	shouldNotMatch("9111", "BR")
	shouldNotMatch("1900", "BR")
	shouldNotMatch("9996", "BR")
}
