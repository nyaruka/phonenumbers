// Port of internal/prefixmapper/* (shared prefix lookup) from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package phonenumbers

import (
	"embed"
	"errors"
	"io/fs"
	"strconv"
	"sync"

	"github.com/nyaruka/phonenumbers/v2/internal/serialize"
)

var (
	dataLoadMutex = sync.Mutex{}

	// These are maps for our prefix to carrier maps
	carrierPrefixMap = make(map[string]*serialize.IntStringMap)

	// These are maps for our prefix to geocoding maps
	geocodingPrefixMap = make(map[string]*serialize.IntStringMap)

	// Our once and map for prefix to timezone lookups
	timezoneOnce sync.Once
	timezoneMap  *serialize.IntStringArrayMap
)

func lazyLoadPrefixes(langMap map[string]*serialize.IntStringMap, dataFS embed.FS, dir, language string) (*serialize.IntStringMap, error) {
	dataLoadMutex.Lock()
	defer dataLoadMutex.Unlock()

	// if we already have prefixes (or nil if they don't exist) return that
	prefixes, ok := langMap[language]
	if ok {
		return prefixes, nil
	}

	// try to load the data file for this language
	data, err := dataFS.ReadFile(dir + "/" + language + ".txt.gz")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	if data != nil {
		prefixes, err = serialize.LoadPrefixMap(data)
		if err != nil {
			return nil, err
		}
		langMap[language] = prefixes
	} else {
		langMap[language] = nil // language doesn't have data
	}

	return langMap[language], nil
}

func getValueForNumber(langMap map[string]*serialize.IntStringMap, dataFS embed.FS, dir, language string, maxLength int, number *PhoneNumber) (string, int, error) {
	prefixes, err := lazyLoadPrefixes(langMap, dataFS, dir, language)
	if err != nil || prefixes == nil {
		return "", 0, err
	}

	e164 := Format(number, E164)

	l := len(e164)
	if maxLength > l {
		maxLength = l
	}
	for i := maxLength; i > 1; i-- {
		index, err := strconv.Atoi(e164[0:i])
		if err != nil {
			return "", 0, err
		}
		if value, has := prefixes.Map[index]; has {
			return value, index, nil
		}
	}
	return "", 0, nil
}
