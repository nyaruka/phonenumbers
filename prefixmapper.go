// Port of internal/prefixmapper/* (shared prefix lookup) from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package phonenumbers

import (
	"embed"
	"errors"
	"io/fs"
	"strconv"
	"sync"
)

var (
	dataLoadMutex = sync.Mutex{}

	// These are maps for our prefix to carrier maps
	carrierPrefixMap = make(map[string]*intStringMap)

	// These are maps for our prefix to geocoding maps
	geocodingPrefixMap = make(map[string]*intStringMap)

	// Our once and map for prefix to timezone lookups
	timezoneOnce sync.Once
	timezoneMap  *intStringArrayMap
)

func lazyLoadPrefixes(langMap map[string]*intStringMap, dataFS embed.FS, dir, language string) (*intStringMap, error) {
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
		prefixes, err = loadPrefixMap(data)
		if err != nil {
			return nil, err
		}
		langMap[language] = prefixes
	} else {
		langMap[language] = nil // language doesn't have data
	}

	return langMap[language], nil
}

func getValueForNumber(langMap map[string]*intStringMap, dataFS embed.FS, dir, language string, maxLength int, number *PhoneNumber) (string, int, error) {
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
