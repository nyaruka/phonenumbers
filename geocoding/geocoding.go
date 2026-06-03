// Port of java/geocoder/src/com/google/i18n/phonenumbers/geocoding/PhoneNumberOfflineGeocoder.java.
package geocoding

import (
	"embed"
	"strings"

	"github.com/nyaruka/phonenumbers/v2"
	"github.com/nyaruka/phonenumbers/v2/internal/prefixmapper"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

//go:embed data/*.gz
var geocodingData embed.FS

var mapper = prefixmapper.New(geocodingData, "data")

// DescriptionForValidNumber returns a text description in the given language for
// the geographical area the number is from, falling back to the country name.
// The number is assumed to be valid.
func DescriptionForValidNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	mobileToken := phonenumbers.GetCountryMobileToken(int(number.GetCountryCode()))
	nationalNumber := phonenumbers.GetNationalSignificantNumber(number)

	lookupNumber := number
	if mobileToken != "" && strings.HasPrefix(nationalNumber, mobileToken) {
		// In some countries, e.g. Argentina, mobile numbers have a mobile token before the
		// national destination code; this should be removed before geocoding.
		region := phonenumbers.GetRegionCodeForCountryCode(int(number.GetCountryCode()))
		if copied, err := phonenumbers.Parse(nationalNumber[len(mobileToken):], region); err == nil {
			lookupNumber = copied
		}
	}

	e164 := phonenumbers.Format(lookupNumber, phonenumbers.E164)
	areaDescription, _, err := mapper.ValueForNumber(lang, 10, e164)
	if err != nil {
		return "", err
	}
	if areaDescription != "" {
		return areaDescription, nil
	}
	return countryNameForNumber(number, lang)
}

// DescriptionForNumber returns a text description in the given language for the
// given phone number: the name of the geographical area it is from for
// geographical numbers, otherwise the name of the country, and the empty string
// for numbers of an unknown type.
func DescriptionForNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	numberType := phonenumbers.GetNumberType(number)
	if numberType == phonenumbers.UNKNOWN {
		return "", nil
	} else if !phonenumbers.IsNumberGeographicalForType(numberType, int(number.GetCountryCode())) {
		return countryNameForNumber(number, lang)
	}
	return DescriptionForValidNumber(number, lang)
}

// countryNameForNumber returns the name of the country, in the given language,
// for the region the number is from, returning the empty string when the number
// is valid for more than one region and a single one cannot be determined.
func countryNameForNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	regionCodes := phonenumbers.GetRegionCodesForCountryCode(int(number.GetCountryCode()))
	if len(regionCodes) == 1 {
		return regionDisplayName(regionCodes[0], lang), nil
	}
	regionWhereValid := "ZZ"
	for _, regionCode := range regionCodes {
		if phonenumbers.IsValidNumberForRegion(number, regionCode) {
			// If the number has already been found valid for one region, then we don't know
			// which region it belongs to so we return nothing.
			if regionWhereValid != "ZZ" {
				return "", nil
			}
			regionWhereValid = regionCode
		}
	}
	return regionDisplayName(regionWhereValid, lang), nil
}

// regionDisplayName returns the localized country name for a region code, or the
// empty string for unknown or non-geographical regions.
func regionDisplayName(regionCode, lang string) string {
	if regionCode == "" || regionCode == "ZZ" || regionCode == phonenumbers.REGION_CODE_FOR_NON_GEO_ENTITY {
		return ""
	}
	reg, err := language.ParseRegion(regionCode)
	if err != nil {
		return ""
	}
	langT, err := language.Parse(lang)
	if err != nil {
		langT = language.English // fall back to english only for rendering the country name
	}
	return display.Regions(langT).Name(reg)
}
