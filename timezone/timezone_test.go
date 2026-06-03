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

func TestForPrefix(t *testing.T) {
	tests := []timeZonesTestCases{
		{
			num:               "+442073238299",
			expectedTimeZones: []string{"Europe/London"},
		},
		{
			num:               "+61491570156",
			expectedTimeZones: []string{"Australia/Sydney"},
		},
		{
			num:               "+61255501234",
			expectedTimeZones: []string{"Australia/Sydney"},
		},
		{
			num:               "+12067798181",
			expectedTimeZones: []string{"America/Los_Angeles"},
		},
		{
			num:               "+390399123456",
			expectedTimeZones: []string{"Europe/Rome"},
		},
		{
			num:               "+541151123456",
			expectedTimeZones: []string{"America/Buenos_Aires"},
		},
		{
			num:               "+15167706076",
			expectedTimeZones: []string{"America/New_York"},
		},
		{
			num:               "+917999999543",
			expectedTimeZones: []string{"Asia/Calcutta"},
		},
		{
			num:               "+540111561234567",
			expectedTimeZones: []string{"America/Buenos_Aires"},
		},
		{
			num:               "+18504320800",
			expectedTimeZones: []string{"America/Chicago"},
		},
		{
			num:               "+14079395277",
			expectedTimeZones: []string{"America/New_York"},
		},
		{
			num:               "+18508632167",
			expectedTimeZones: []string{"America/Chicago"},
		},
		{
			num:               "+40213158207",
			expectedTimeZones: []string{"Europe/Bucharest"},
		},
		// UTC +5:45
		{
			num:               "+97714240520",
			expectedTimeZones: []string{"Asia/Katmandu"},
		},
		// UTC -3:30
		{
			num:               "+17097264534",
			expectedTimeZones: []string{"America/St_Johns"},
		},
		{
			num:               "0000000000",
			expectedTimeZones: []string{Unknown},
		},
		{
			num:               "+31112",
			expectedTimeZones: []string{"Europe/Amsterdam"},
		},
		{
			num:               "+6837000",
			expectedTimeZones: []string{"Pacific/Niue"},
		},
		{
			num: "+1911",
			expectedTimeZones: []string{
				"America/Adak", "America/Anchorage", "America/Anguilla", "America/Antigua",
				"America/Barbados", "America/Boise", "America/Cayman", "America/Chicago",
				"America/Denver", "America/Dominica", "America/Edmonton", "America/Fort_Nelson",
				"America/Grand_Turk", "America/Grenada", "America/Halifax", "America/Jamaica",
				"America/Juneau", "America/Los_Angeles", "America/Lower_Princes", "America/Montserrat",
				"America/Nassau", "America/New_York", "America/North_Dakota/Center",
				"America/Phoenix", "America/Port_of_Spain", "America/Puerto_Rico",
				"America/Regina", "America/Santo_Domingo", "America/St_Johns", "America/St_Kitts",
				"America/St_Lucia", "America/St_Thomas", "America/St_Vincent", "America/Toronto",
				"America/Tortola", "America/Vancouver", "America/Winnipeg", "Atlantic/Bermuda",
				"Pacific/Guam", "Pacific/Honolulu", "Pacific/Pago_Pago", "Pacific/Saipan",
			},
		},
	}

	for _, test := range tests {
		timeZones, err := ForPrefix(test.num)
		if err != nil {
			t.Errorf("Failed to getTimezone for the number %s: %s", test.num, err)
		}

		if len(timeZones) == 0 {
			t.Errorf("Expected at least 1 timezone.")
		}

		if !reflect.DeepEqual(timeZones, test.expectedTimeZones) {
			t.Errorf("Expected '%v', got '%v' for '%s'", test.expectedTimeZones, timeZones, test.num)
		}

		num, err := phonenumbers.Parse(test.num, "")
		if err != nil {
			continue
		}

		timeZones, err = ForNumber(num)
		if err != nil {
			t.Errorf("Failed to getTimezone for the number %s: %s", num, err)
		}

		if len(timeZones) == 0 {
			t.Errorf("Expected at least 1 timezone.")
		}

		if !reflect.DeepEqual(timeZones, test.expectedTimeZones) {
			t.Errorf("Expected '%v', got '%v' for '%s'", test.expectedTimeZones, timeZones, num)
		}
	}
}
