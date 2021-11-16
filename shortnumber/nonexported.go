package shortnumber

// Contains exact copies of non-exporterd things from the phonenumbers package

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"regexp"

	"github.com/nyaruka/phonenumbers"
)

func extractPossibleNumber(number string) string {
	if phonenumbers.VALID_START_CHAR_PATTERN.MatchString(number) {
		start := phonenumbers.VALID_START_CHAR_PATTERN.FindIndex([]byte(number))[0]
		number = number[start:]
		// Remove trailing non-alpha non-numerical characters.
		indices := phonenumbers.UNWANTED_END_CHAR_PATTERN.FindIndex([]byte(number))
		if len(indices) > 0 {
			number = number[0:indices[0]]
		}
		// Check for extra numbers at the end.
		indices = phonenumbers.SECOND_NUMBER_START_PATTERN.FindIndex([]byte(number))
		if len(indices) > 0 {
			number = number[0:indices[0]]
		}
		return number
	}
	return ""
}

func readFromRegexCache(key string) (*regexp.Regexp, bool) {
	regCacheMutex.RLock()
	v, ok := regexCache[key]
	regCacheMutex.RUnlock()
	return v, ok
}

func regexFor(pattern string) *regexp.Regexp {
	regex, found := readFromRegexCache(pattern)
	if !found {
		regex = regexp.MustCompile(pattern)
		writeToRegexCache(pattern, regex)
	}
	return regex
}

func writeToRegexCache(key string, value *regexp.Regexp) {
	regCacheMutex.Lock()
	regexCache[key] = value
	regCacheMutex.Unlock()
}

func decodeUnzipString(data string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	zipReader, err := gzip.NewReader(bytes.NewReader(decodedBytes))
	if err != nil {
		return nil, err
	}

	rawBytes, err := ioutil.ReadAll(zipReader)
	if err != nil {
		return nil, err
	}

	return rawBytes, nil
}

// intStringArrayMap is our map from an int to an array of strings
// this is used for our timezone and region maps
type intStringArrayMap struct {
	Map       map[int][]string
	MaxLength int
}

// intStringMap is our data structure for maps from prefixes to a single string
// this is used for our carrier and geocoding maps
type intStringMap struct {
	Map       map[int]string
	MaxLength int
}
