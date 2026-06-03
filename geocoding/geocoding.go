// Port of java/geocoder/src/com/google/i18n/phonenumbers/geocoding/PhoneNumberOfflineGeocoder.java.
// Functions are kept in upstream source order to ease syncing.
package geocoding

import (
	"embed"

	"github.com/nyaruka/phonenumbers/v2"
	"github.com/nyaruka/phonenumbers/v2/internal/prefixmapper"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

//go:embed data/*.gz
var geocodingData embed.FS

var mapper = prefixmapper.New(geocodingData, "data")

// ForNumber returns the location we think the number was first acquired in. This is
// just our best guess, there is no guarantee to its accuracy.
func ForNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	e164 := phonenumbers.Format(number, phonenumbers.E164)

	geocoding, _, err := mapper.ValueForNumber(lang, 10, e164)
	if err != nil || geocoding != "" {
		return geocoding, err
	}

	// fallback to english
	geocoding, _, err = mapper.ValueForNumber("en", 10, e164)
	if err != nil || geocoding != "" {
		return geocoding, err
	}

	// fallback to locale
	var reg language.Region
	if reg, err = language.ParseRegion(phonenumbers.GetRegionCodeForNumber(number)); err != nil {
		return "", err
	}

	var langT language.Tag
	if langT, err = language.Parse(lang); err != nil {
		langT = language.English // fallback to english
	}
	return display.Regions(langT).Name(reg), nil
}
