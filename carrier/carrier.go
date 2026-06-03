// Port of carrier/PhoneNumberToCarrierMapper.java from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package carrier

import (
	"embed"

	"github.com/nyaruka/phonenumbers/v2"
	"github.com/nyaruka/phonenumbers/v2/internal/prefixmapper"
)

//go:embed data/*.gz
var carrierData embed.FS

var mapper = prefixmapper.New(carrierData, "data")

// ForNumber returns the carrier we believe the number belongs to. Note due
// to number porting this is only a guess, there is no guarantee to its accuracy.
func ForNumber(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	carrier, _, err := WithPrefixForNumber(number, lang)
	return carrier, err
}

// SafeDisplayName gets the name of the carrier for the given phone number
// only when it is 'safe' to display to users.
// A carrier name is considered safe if the number is valid and
// for a region that doesn't support mobile number portability .
func SafeDisplayName(number *phonenumbers.PhoneNumber, lang string) (string, error) {
	if phonenumbers.IsMobileNumberPortableRegion(phonenumbers.GetRegionCodeForNumber(number)) {
		return "", nil
	}
	return ForNumber(number, lang)
}

// WithPrefixForNumber returns the carrier we believe the number belongs to, as well as
// its prefix. Note due to number porting this is only a guess, there is no guarantee to its accuracy.
func WithPrefixForNumber(number *phonenumbers.PhoneNumber, lang string) (string, int, error) {
	e164 := phonenumbers.Format(number, phonenumbers.E164)

	carrier, prefix, err := mapper.ValueForNumber(lang, 10, e164)
	if err != nil {
		return "", 0, err
	}
	if carrier != "" {
		return carrier, prefix, nil
	}

	// fallback to english
	return mapper.ValueForNumber("en", 10, e164)
}
