// Port of java/geocoder/src/com/google/i18n/phonenumbers/PhoneNumberToTimeZonesMapper.java.
package timezone

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/nyaruka/phonenumbers/v2"
	"github.com/nyaruka/phonenumbers/v2/internal/serialize"
)

//go:embed data/prefix_to_timezone.xml.gz
var timezoneData []byte

// Unknown is the value returned for a prefix that maps to no timezone, defined
// by ICU as the unknown time zone. It corresponds to upstream's
// PhoneNumberToTimeZonesMapper.getUnknownTimeZone(), exposed here as a const
// rather than a function.
const Unknown = "Etc/Unknown"

var (
	timezoneOnce sync.Once
	timezoneMap  *serialize.IntStringArrayMap
)

// GetTimeZonesForNumber returns the names of the timezones to which the given
// number maps. An unknown-type number maps to the unknown timezone; a
// non-geographical number is resolved at the country-calling-code level; any
// other number is resolved from its full digits.
func GetTimeZonesForNumber(number *phonenumbers.PhoneNumber) ([]string, error) {
	numberType := phonenumbers.GetNumberType(number)
	if numberType == phonenumbers.UNKNOWN {
		return []string{Unknown}, nil
	} else if !phonenumbers.IsNumberGeographicalForType(numberType, int(number.GetCountryCode())) {
		return countryLevelTimeZonesForNumber(number)
	}
	return GetTimeZonesForGeographicalNumber(number)
}

// GetTimeZonesForGeographicalNumber returns the names of the timezones to which the
// given geographical number maps, resolved from its full digits.
func GetTimeZonesForGeographicalNumber(number *phonenumbers.PhoneNumber) ([]string, error) {
	e164 := phonenumbers.Format(number, phonenumbers.E164)
	return lookupTimeZones(e164)
}

// countryLevelTimeZonesForNumber resolves the timezones for a number using only
// its country calling code.
func countryLevelTimeZonesForNumber(number *phonenumbers.PhoneNumber) ([]string, error) {
	return lookupTimeZones(strconv.Itoa(int(number.GetCountryCode())))
}

// lookupTimeZones returns the timezones for the longest matching prefix of the
// given number string, or the unknown timezone when none matches. The algorithm
// tries to match starting from the maximum number of digits and decreasing until
// it finds one or reaches 0.
func lookupTimeZones(number string) ([]string, error) {
	var err error
	timezoneOnce.Do(func() {
		timezoneMap, err = serialize.LoadIntArrayMap(timezoneData)
	})

	if timezoneMap == nil {
		return nil, fmt.Errorf("error loading timezone map: %v", err)
	}

	// strip any leading +
	number = strings.TrimLeft(number, "+")

	matchLength := min(len(number), timezoneMap.MaxLength)

	for i := matchLength; i > 0; i-- {
		index, err := strconv.Atoi(number[0:i])
		if err != nil {
			return nil, err
		}
		tzs, found := timezoneMap.Map[index]
		if found {
			return tzs, nil
		}
	}
	return []string{Unknown}, nil
}
