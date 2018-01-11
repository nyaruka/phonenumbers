package phonenumbers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
)

type prefixData struct {
	prefixToValue   map[int]map[string]string
	maxPrefixLength int
}

var carrierInstance, geocodingInstance *prefixData
var carrierOnce, geocodingOnce sync.Once

func loadData(srcData map[string]string) *prefixData {
	instance := &prefixData{make(map[int]map[string]string), 0}
	for file, body := range srcData {
		items := strings.Split(file, "/")
		if len(items) != 3 {
			continue
		}

		data, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			continue
		}
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			continue
		}
		rawBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(rawBytes), "\n") {
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			fields := strings.Split(line, "|")
			if len(fields) != 2 {
				continue
			}

			if len(fields[0]) > instance.maxPrefixLength {
				instance.maxPrefixLength = len(fields[0])
			}
			code, err := strconv.Atoi(fields[0])
			if err != nil {
				continue
			}
			if c, has := instance.prefixToValue[code]; has {
				c[items[1]] = fields[1]
			} else {
				instance.prefixToValue[code] =
					map[string]string{items[1]: fields[1]}
			}
		}
	}
	return instance
}

func getData(src string) *prefixData {
	switch src {
	case "carrier":
		carrierOnce.Do(func() {
			carrierInstance = loadData(Carriers)
		})
		return carrierInstance
	case "geocoding":
		geocodingOnce.Do(func() {
			geocodingInstance = loadData(Geocodings)
		})
		return geocodingInstance
	}
	return nil
}

func getValueForNumber(src string, number *PhoneNumber, lang string) (string, error) {
	e164 := Format(number, E164)
	instance := getData(src)
	max := instance.maxPrefixLength + 1
	if max > len(e164) {
		max = len(e164)
	}
	for i := max; i > 1; i-- {
		index, err := strconv.Atoi(e164[0:i])
		if err != nil {
			return "", err
		}
		if c, has := instance.prefixToValue[index]; has {
			if c[lang] != "" {
				return c[lang], nil
			}
			// fall back to English
			return c["en"], nil
		}
	}
	return "", nil
}

func GetCarrierForNumber(number *PhoneNumber, lang string) (string, error) {
	return getValueForNumber("carrier", number, lang)
}

func GetGeocodingForNumber(number *PhoneNumber, lang string) (string, error) {
	return getValueForNumber("geocoding", number, lang)
}
