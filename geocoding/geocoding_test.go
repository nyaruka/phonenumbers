package geocoding

import (
	"testing"

	"github.com/nyaruka/phonenumbers/v2"
)

func TestGetDescriptionForNumber(t *testing.T) {
	tests := []struct {
		num      string
		lang     string
		expected string
	}{
		{num: "+8613702032331", lang: "en", expected: "Tianjin"},
		{num: "+8613702032331", lang: "zh", expected: "天津市"},
		{num: "+863197785050", lang: "zh", expected: "河北省邢台市"},
		{num: "+8613323241342", lang: "en", expected: "Baoding, Hebei"},
		{num: "+917999499543", lang: "en", expected: "Ahmedabad Local, Gujarat"},
		{num: "+17047181840", lang: "en", expected: "North Carolina"},
		{num: "+12542462158", lang: "en", expected: "Texas"},
		{num: "+16193165996", lang: "en", expected: "California"},
		{num: "+12067799191", lang: "en", expected: "Washington State"},
		{num: "+447825602614", lang: "en", expected: "United Kingdom"},
	}
	for _, test := range tests {
		number, err := phonenumbers.Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		geocoding, err := GetDescriptionForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to get description for the number %s: %s", test.num, err)
		}
		if test.expected != geocoding {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expected, geocoding, test.num)
		}
	}
}

func TestGetDescriptionForNumberForUserRegion(t *testing.T) {
	tests := []struct {
		num        string
		lang       string
		userRegion string
		expected   string
	}{
		// User in the same region as the number: detailed area description.
		{num: "+16193165996", lang: "en", userRegion: "US", expected: "California"},
		// User in a different region: just the country name, in the requested language,
		// without the more detailed area information.
		{num: "+16193165996", lang: "en", userRegion: "GB", expected: "United States"},
		{num: "+16193165996", lang: "es", userRegion: "IT", expected: "Estados Unidos"},
		// Unknown user region: just show the country name.
		{num: "+16193165996", lang: "es", userRegion: "ZZ", expected: "Estados Unidos"},
		// Invalid number: empty string.
		{num: "+1123456789", lang: "en", userRegion: "US", expected: ""},
	}
	for _, test := range tests {
		number, err := phonenumbers.Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		geocoding, err := GetDescriptionForNumberForUserRegion(number, test.lang, test.userRegion)
		if err != nil {
			t.Errorf("Failed to get description for the number %s: %s", test.num, err)
		}
		if test.expected != geocoding {
			t.Errorf("Expected '%s', got '%s' for '%s' (userRegion %s)", test.expected, geocoding, test.num, test.userRegion)
		}
	}
}
