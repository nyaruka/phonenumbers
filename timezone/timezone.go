// Port of java/geocoder/src/com/google/i18n/phonenumbers/PhoneNumberToTimeZonesMapper.java from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
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

// Unknown is the value returned for a prefix that maps to no timezone. This is
// defined by ICU as the unknown time zone.
const Unknown = "Etc/Unknown"

var (
	timezoneOnce sync.Once
	timezoneMap  *serialize.IntStringArrayMap
)

// ForPrefix returns a slice of Timezones corresponding to the number passed
// or error when it is impossible to convert the string to int
// The algorithm tries to match the timezones starting from the maximum
// number of phone number digits and decreasing until it finds one or reaches 0
func ForPrefix(number string) ([]string, error) {
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

// ForNumber returns the names of timezones which we believe maps to the
// passed in number.
func ForNumber(number *phonenumbers.PhoneNumber) ([]string, error) {
	e164 := phonenumbers.Format(number, phonenumbers.E164)
	return ForPrefix(e164)
}
