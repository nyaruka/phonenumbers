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

// GetDescriptionForValidNumber returns a text description in the given language for
// the geographical area the number is from, falling back to the country name.
// The number is assumed to be valid.
func GetDescriptionForValidNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
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

// GetDescriptionForValidNumberForUserRegion is as per GetDescriptionForValidNumber but
// also considers the region of the user. If the phone number is from the same region as
// the user, only a lower-level description will be returned, if one exists. Otherwise,
// the phone number's region will be returned, with optionally some more detailed
// information.
//
// For example, for a user from the region "US" (United States), we would show "Mountain
// View, CA" for a particular number, omitting the United States from the description. For
// a user from the United Kingdom (region "GB"), for the same number we may show "Mountain
// View, CA, United States" or even just "United States". The number is assumed to be
// valid. userRegion should be a two-letter upper-case CLDR region code.
func GetDescriptionForValidNumberForUserRegion(number *phonenumbers.PhoneNumber, lang string, userRegion string) (string, error) {
	// If the user region matches the number's region, then we just show the lower-level
	// description, if one exists - if no description exists, we will show the region(country)
	// name for the number.
	regionCode := phonenumbers.GetRegionCodeForNumber(number)
	if userRegion == regionCode {
		return GetDescriptionForValidNumber(number, lang)
	}
	// Otherwise, we just show the region(country) name for now.
	return regionDisplayName(regionCode, lang), nil
	// TODO: Concatenate the lower-level and country-name information in an appropriate
	// way for each language.
}

// GetDescriptionForNumber returns a text description in the given language for the
// given phone number: the name of the geographical area it is from for
// geographical numbers, otherwise the name of the country, and the empty string
// for numbers of an unknown type.
func GetDescriptionForNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	numberType := phonenumbers.GetNumberType(number)
	if numberType == phonenumbers.UNKNOWN {
		return "", nil
	} else if !phonenumbers.IsNumberGeographicalForType(numberType, int(number.GetCountryCode())) {
		return countryNameForNumber(number, lang)
	}
	return GetDescriptionForValidNumber(number, lang)
}

// GetDescriptionForNumberForUserRegion is as per GetDescriptionForValidNumberForUserRegion
// but explicitly checks the validity of the number passed in.
func GetDescriptionForNumberForUserRegion(number *phonenumbers.PhoneNumber, lang string, userRegion string) (string, error) {
	numberType := phonenumbers.GetNumberType(number)
	if numberType == phonenumbers.UNKNOWN {
		return "", nil
	} else if !phonenumbers.IsNumberGeographicalForType(numberType, int(number.GetCountryCode())) {
		return countryNameForNumber(number, lang)
	}
	return GetDescriptionForValidNumberForUserRegion(number, lang, userRegion)
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
