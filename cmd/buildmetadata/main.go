package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"bytes"

	"github.com/gogo/protobuf/proto"
	"github.com/nyaruka/gophone"
)

const (
	metadataURL  = `https://raw.githubusercontent.com/googlei18n/libphonenumber/master/resources/PhoneNumberMetadata.xml`
	metadataPath = `src/github.com/nyaruka/gophone/metadata_bin.go`

	tzURL  = `https://raw.githubusercontent.com/googlei18n/libphonenumber/master/resources/timezones/map_data.txt`
	tzPath = `src/github.com/nyaruka/gophone/prefix_to_timezone.go`

	regionPath = `src/github.com/nyaruka/gophone/countrycode_to_region.go`
)

func fetchURL(url string) []byte {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("Error fetching URL '%s': %s", url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading body: %s", err)
	}

	return body
}

func writeFile(filePath string, data []byte) {
	gopath, found := os.LookupEnv("GOPATH")
	if !found {
		log.Fatal("Missing $GOPATH environment variable")
	}

	path := filepath.Join(gopath, filePath)

	err := ioutil.WriteFile(path, data, os.FileMode(0774))
	if err != nil {
		log.Fatalf("Error writing '%s': %s", path, err)
	}
}

func buildRegions(metadata *gophone.PhoneMetadataCollection) {
	log.Println("Building region map")
	regionMap := gophone.BuildCountryCodeToRegionMap(metadata)

	output := bytes.Buffer{}
	output.WriteString("package gophone\n\n")
	output.WriteString("// CountryCodeToRegion maps a country code to a list of possible regions\n")
	output.WriteString("var CountryCodeToRegion = map[int][]string{\n")

	ccs := make([]int, 0, len(regionMap))
	for cc := range regionMap {
		ccs = append(ccs, int(cc))
	}
	sort.Ints(ccs)

	for _, cc := range ccs {
		regions := regionMap[int32(cc)]

		// write our map entry
		output.WriteString("\t")
		output.WriteString(strconv.FormatInt(int64(cc), 10))
		output.WriteString(":")
		if cc < 10 {
			output.WriteString("  ")
		} else if cc < 100 {
			output.WriteString(" ")
		}
		output.WriteString(" []string{")

		if len(regions) == 1 {
			output.WriteString("\"")
			output.WriteString(regions[0])
			output.WriteString("\"},\n")
		} else {
			sort.Strings(regions[1:])
			for i, region := range regions {
				if i%10 == 0 {
					output.WriteString("\n\t\t")
				}
				output.WriteString("\"")
				output.WriteString(region)
				output.WriteString("\", ")
			}

			output.WriteString("\n\t},\n")
		}
	}

	output.WriteString("\n}\n")

	log.Println("Writing new countrycode_to_region.go")
	writeFile(regionPath, output.Bytes())
}

func buildTimezones() {
	log.Println("Fetching map_data.txt from Github")
	body := fetchURL(tzURL)

	// create our output
	output := bytes.Buffer{}
	output.WriteString("package gophone\n\n")
	output.WriteString("// PrefixToTimezone maps a phone number prefix to a list of possible timezones\n")
	output.WriteString("var PrefixToTimezone = map[int][]string{\n")

	// we keep track of the max prefix length we have so we can write it afterwards
	maxPrefixLength := 0

	// read through our body
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) != 2 {
			log.Fatalf("Invalid format in timezone file: %s", line)
		}

		zones := strings.Split(fields[1], "&")
		if len(zones) < 1 {
			log.Fatalf("Invalid format in timezone file: %s", line)
		}

		// write our map entry
		output.WriteString("\t")
		output.WriteString(fields[0])
		output.WriteString(": []string{")

		if len(fields[0]) > maxPrefixLength {
			maxPrefixLength = len(fields[0])
		}

		if len(zones) == 1 {
			output.WriteString("\"")
			output.WriteString(zones[0])
			output.WriteString("\"},\n")
		} else {
			sort.Strings(zones)
			for i, zone := range zones {
				if i%5 == 0 {
					output.WriteString("\n\t\t")
				}
				output.WriteString("\"")
				output.WriteString(zone)
				output.WriteString("\", ")
			}

			output.WriteString("\n\t},\n")
		}
	}
	output.WriteString("\n}\n")

	output.WriteString("// MAX_PREFIX_LENGTH helps us optimize for prefixes that can't be contained in our table\n")
	output.WriteString("var MAX_PREFIX_LENGTH = ")
	output.WriteString(strconv.FormatInt(int64(maxPrefixLength), 10))
	output.WriteString("\n")

	log.Println("Writing new prefix_to_timezone.go")
	writeFile(tzPath, output.Bytes())
}

func buildMetadata() *gophone.PhoneMetadataCollection {
	log.Println("Fetching PhoneNumberMetadata.xml from Github")
	body := fetchURL(metadataURL)

	log.Println("Building new metadata collection")
	collection, err := gophone.BuildPhoneMetadataCollection(body, false, false)
	if err != nil {
		log.Fatalf("Error converting XML: %s", err)
	}

	// now that we've generated our possible patterns we can get rid of possible lengths in our proto buffers
	for _, md := range collection.Metadata {
		md.ClearPossibleLengths()
	}

	// write it out as a protobuf
	log.Println("Writing new metadata_bin.go")
	data, err := proto.Marshal(collection)
	if err != nil {
		log.Fatalf("Error marshalling metadata: %v", err)
	}

	// zip it
	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	w.Write(data)
	w.Close()
	data = compressed.Bytes()

	// create our output
	output := bytes.Buffer{}

	// write our header
	output.WriteString("package gophone\n\n")
	output.WriteString("var MetaData = []byte{\n        ")

	// ok, write each byte of our data in groups of 14
	for i, b := range data {
		if i > 0 && i%14 == 0 {
			output.WriteString("\n        ")
		}
		output.WriteString("0x")
		output.WriteString(fmt.Sprintf("%02x", b))
		output.WriteString(", ")
	}

	output.WriteString("\n}\n")

	log.Println("Writing new metadata_bin.go")
	writeFile(metadataPath, output.Bytes())

	return collection
}

func main() {
	metadata := buildMetadata()
	buildRegions(metadata)
	buildTimezones()
}
