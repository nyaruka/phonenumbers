// Port of java/carrier/src/com/google/i18n/phonenumbers/PhoneNumberToCarrierMapper.java.
package carrier

import (
	"embed"

	"github.com/nyaruka/phonenumbers/v2"
	"github.com/nyaruka/phonenumbers/v2/internal/prefixmapper"
)

//go:embed data/*.gz
var carrierData embed.FS

var mapper = prefixmapper.New(carrierData, "data")

// GetNameForValidNumber returns the carrier we believe the number belongs to. Note
// due to number porting this is only a guess; there is no guarantee of its
// accuracy. The number is assumed to be valid.
func GetNameForValidNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	e164 := phonenumbers.Format(number, phonenumbers.E164)
	name, _, err := mapper.ValueForNumber(lang, 10, e164)
	return name, err
}

// GetNameForNumber returns the carrier we believe the number belongs to. Note due
// to number porting this is only a guess; there is no guarantee of its accuracy.
// A carrier name is only returned for numbers of a mobile type.
func GetNameForNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	if isMobile(phonenumbers.GetNumberType(number)) {
		return GetNameForValidNumber(number, lang)
	}
	return "", nil
}

// GetSafeDisplayName gets the name of the carrier for the given phone number only
// when it is 'safe' to display to users. A carrier name is considered safe if
// the number is valid and for a region that doesn't support mobile number
// portability.
func GetSafeDisplayName(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	if phonenumbers.IsMobileNumberPortableRegion(phonenumbers.GetRegionCodeForNumber(number)) {
		return "", nil
	}
	return GetNameForNumber(number, lang)
}

// isMobile reports whether numberType is a mobile type that a carrier can be
// looked up for.
func isMobile(numberType phonenumbers.PhoneNumberType) bool {
	return numberType == phonenumbers.MOBILE ||
		numberType == phonenumbers.FIXED_LINE_OR_MOBILE ||
		numberType == phonenumbers.PAGER
}
