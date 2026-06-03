package timezone

import (
	"reflect"
	"testing"

	"github.com/nyaruka/phonenumbers/v2"
)

type timeZonesTestCases struct {
	num               string
	expectedTimeZones []string
}

func TestTimeZonesForNumber(t *testing.T) {
	tests := []timeZonesTestCases{
		{num: "+442073238299", expectedTimeZones: []string{"Europe/London"}},
		{num: "+61255501234", expectedTimeZones: []string{"Australia/Sydney"}},
		{num: "+12067798181", expectedTimeZones: []string{"America/Los_Angeles"}},
		{num: "+390399123456", expectedTimeZones: []string{"Europe/Rome"}},
		{num: "+541151123456", expectedTimeZones: []string{"America/Buenos_Aires"}},
		{num: "+15167706076", expectedTimeZones: []string{"America/New_York"}},
		{num: "+917999999543", expectedTimeZones: []string{"Asia/Calcutta"}},
		{num: "+540111561234567", expectedTimeZones: []string{"America/Buenos_Aires"}},
		{num: "+18504320800", expectedTimeZones: []string{"America/Chicago"}},
		{num: "+14079395277", expectedTimeZones: []string{"America/New_York"}},
		{num: "+18508632167", expectedTimeZones: []string{"America/Chicago"}},
		{num: "+40213158207", expectedTimeZones: []string{"Europe/Bucharest"}},
		// UTC +5:45
		{num: "+97714240520", expectedTimeZones: []string{"Asia/Katmandu"}},
		// UTC -3:30
		{num: "+17097264534", expectedTimeZones: []string{"America/St_Johns"}},
		{num: "+6837000", expectedTimeZones: []string{"Pacific/Niue"}},
		// A mobile number in a country where mobile numbers aren't geographically
		// assigned is non-geographical, so it resolves at the country-calling-code
		// level to all of the country's timezones.
		{num: "+61491570156", expectedTimeZones: []string{
			"Australia/Adelaide", "Australia/Brisbane", "Australia/Eucla", "Australia/Lord_Howe",
			"Australia/Perth", "Australia/Sydney", "Indian/Christmas", "Indian/Cocos",
		}},
	}

	for _, test := range tests {
		num, err := phonenumbers.Parse(test.num, "")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
			continue
		}

		timeZones, err := TimeZonesForNumber(num)
		if err != nil {
			t.Errorf("Failed to get timezones for the number %s: %s", test.num, err)
		}
		if len(timeZones) == 0 {
			t.Errorf("Expected at least 1 timezone for %s.", test.num)
		}
		if !reflect.DeepEqual(timeZones, test.expectedTimeZones) {
			t.Errorf("Expected '%v', got '%v' for '%s'", test.expectedTimeZones, timeZones, test.num)
		}
	}
}

// A number whose type can't be determined maps to the unknown timezone.
func TestTimeZonesForNumberUnknownType(t *testing.T) {
	num, err := phonenumbers.Parse("+1911", "")
	if err != nil {
		t.Fatalf("Failed to parse number: %s", err)
	}
	timeZones, err := TimeZonesForNumber(num)
	if err != nil {
		t.Fatalf("Failed to get timezones: %s", err)
	}
	if !reflect.DeepEqual(timeZones, []string{Unknown}) {
		t.Errorf("Expected '%v', got '%v'", []string{Unknown}, timeZones)
	}
}

// For a non-geographical mobile, the geographical lookup still resolves to the
// specific area's timezone, unlike TimeZonesForNumber which gives the
// country-level list.
func TestTimeZonesForGeographicalNumber(t *testing.T) {
	num, err := phonenumbers.Parse("+61491570156", "")
	if err != nil {
		t.Fatalf("Failed to parse number: %s", err)
	}
	timeZones, err := TimeZonesForGeographicalNumber(num)
	if err != nil {
		t.Fatalf("Failed to get timezones: %s", err)
	}
	if !reflect.DeepEqual(timeZones, []string{"Australia/Sydney"}) {
		t.Errorf("Expected '%v', got '%v'", []string{"Australia/Sydney"}, timeZones)
	}
}
