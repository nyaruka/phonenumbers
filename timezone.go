// Port of geocoder/PhoneNumberToTimeZonesMapper.java from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package phonenumbers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers/v2/internal/serialize"
)

// GetTimezonesForPrefix returns a slice of Timezones corresponding to the number passed
// or error when it is impossible to convert the string to int
// The algorithm tries to match the timezones starting from the maximum
// number of phone number digits and decreasing until it finds one or reaches 0
func GetTimezonesForPrefix(number string) ([]string, error) {
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
	return []string{UNKNOWN_TIMEZONE}, nil
}

// GetTimezonesForNumber returns the names of timezones which we believe maps to the
// passed in number.
func GetTimezonesForNumber(number *PhoneNumber) ([]string, error) {
	e164 := Format(number, E164)
	return GetTimezonesForPrefix(e164)
}
