// Port of geocoder/geocoding/PhoneNumberOfflineGeocoder.java from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package phonenumbers

import (
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// GetGeocodingForNumber returns the location we think the number was first acquired in. This is
// just our best guess, there is no guarantee to its accuracy.
func GetGeocodingForNumber(number *PhoneNumber, lang string) (string, error) {
	geocoding, _, err := getValueForNumber(geocodingPrefixMap, geocodingData, geocodingDataPath, lang, 10, number)
	if err != nil || geocoding != "" {
		return geocoding, err
	}

	// fallback to english
	geocoding, _, err = getValueForNumber(geocodingPrefixMap, geocodingData, geocodingDataPath, "en", 10, number)
	if err != nil || geocoding != "" {
		return geocoding, err
	}

	// fallback to locale
	var reg language.Region
	if reg, err = language.ParseRegion(GetRegionCodeForNumber(number)); err != nil {
		return "", err
	}

	var langT language.Tag
	if langT, err = language.Parse(lang); err != nil {
		langT = language.English // fallback to english
	}
	return display.Regions(langT).Name(reg), nil
}
