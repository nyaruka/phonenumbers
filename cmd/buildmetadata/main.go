package main

import (
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"bytes"

	"github.com/gogo/protobuf/proto"
	"github.com/nyaruka/phonenumbers"
)

type prefixBuild struct {
	url     string
	dir     string
	srcPath string
	varName string
}

const (
	metadataURL  = "https://raw.githubusercontent.com/googlei18n/libphonenumber/master/resources/PhoneNumberMetadata.xml"
	metadataPath = "src/github.com/nyaruka/phonenumbers/metadata_bin.go"

	tzURL  = "https://raw.githubusercontent.com/googlei18n/libphonenumber/master/resources/timezones/map_data.txt"
	tzPath = "src/github.com/nyaruka/phonenumbers/prefix_to_timezone_bin.go"
	tzVar  = "timezoneMapData"

	regionPath = "src/github.com/nyaruka/phonenumbers/countrycode_to_region_bin.go"
	regionVar  = "regionMapData"
)

var carrier = prefixBuild{
	url:     "https://github.com/googlei18n/libphonenumber/trunk/resources/carrier",
	dir:     "carrier",
	srcPath: "src/github.com/nyaruka/phonenumbers/prefix_to_carriers_bin.go",
	varName: "carrierMapData",
}

var geocoding = prefixBuild{
	url:     "https://github.com/googlei18n/libphonenumber/trunk/resources/geocoding",
	dir:     "geocoding",
	srcPath: "src/github.com/nyaruka/phonenumbers/prefix_to_geocodings_bin.go",
	varName: "geocodingMapData",
}

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

func svnExport(dir string, url string) {
	os.RemoveAll(dir)
	cmd := exec.Command(
		"/bin/bash",
		"-c",
		fmt.Sprintf("svn export %s --force", url),
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Fatal(err, string(data))
	}
	outputBuf := bufio.NewReader(stdout)

	for {
		output, _, err := outputBuf.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		log.Println(string(output))
	}

	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func writeFile(filePath string, data []byte) {
	log.Printf("Writing new %s", filePath)
	gopath, found := os.LookupEnv("GOPATH")
	if !found {
		log.Fatal("Missing $GOPATH environment variable")
	}

	path := filepath.Join(gopath, filePath)

	err := ioutil.WriteFile(path, data, os.FileMode(0664))
	if err != nil {
		log.Fatalf("Error writing '%s': %s", path, err)
	}
}

func buildRegions(metadata *phonenumbers.PhoneMetadataCollection) {
	log.Println("Building region map")
	regionMap := phonenumbers.BuildCountryCodeToRegionMap(metadata)
	writeIntStringArrayMap(regionPath, regionVar, regionMap)
}

func buildTimezones() {
	log.Println("Building timezone map")
	body := fetchURL(tzURL)

	// build our map of prefix to timezones
	prefixMap := make(map[int][]string)
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

		// parse our prefix
		prefix, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Fatalf("Invalid prefix in line: %s", line)
		}
		prefixMap[prefix] = zones
	}

	// then write our file
	writeIntStringArrayMap(tzPath, tzVar, prefixMap)
}

func writeIntStringArrayMap(path string, varName string, prefixMap map[int][]string) {
	// build lists of our keys and values
	keys := make([]int, 0, len(prefixMap))
	values := make([]string, 0, 255)
	seenValues := make(map[string]bool, 255)

	for k, vs := range prefixMap {
		keys = append(keys, k)
		for _, v := range vs {
			_, seen := seenValues[v]
			if !seen {
				seenValues[v] = true
				values = append(values, v)
			}
		}
	}
	sort.Strings(values)
	sort.Ints(keys)

	internMap := make(map[string]int, len(values))
	for i, v := range values {
		internMap[v] = i
	}

	data := &bytes.Buffer{}

	// first write our values, as length of string and raw bytes
	joinedValues := strings.Join(values, "\n")
	if err := binary.Write(data, binary.LittleEndian, uint32(len(joinedValues))); err != nil {
		log.Fatal(err)
	}
	if err := binary.Write(data, binary.LittleEndian, []byte(joinedValues)); err != nil {
		log.Fatal(err)
	}

	// then the number of keys
	if err := binary.Write(data, binary.LittleEndian, uint32(len(keys))); err != nil {
		log.Fatal(err)
	}

	// we write our key / value pairs as a varint of the difference of the previous prefix
	// and a uint16 of the value index
	last := 0
	intBuf := make([]byte, 6, 6)
	for _, key := range keys {
		// first write our prefix
		diff := key - last
		l := binary.PutUvarint(intBuf, uint64(diff))
		if err := binary.Write(data, binary.LittleEndian, intBuf[:l]); err != nil {
			log.Fatal(err)
		}

		// then our values
		values := prefixMap[key]

		// write our number of values
		if err := binary.Write(data, binary.LittleEndian, uint8(len(values))); err != nil {
			log.Fatal(err)
		}

		// then each value as the interned index
		for _, v := range values {
			valueIntern := internMap[v]
			if err := binary.Write(data, binary.LittleEndian, uint16(valueIntern)); err != nil {
				log.Fatal(err)
			}
		}

		last = key
	}

	// then write our file
	writeFile(path, generateBinFile(varName, data.Bytes()))
}

func buildMetadata() *phonenumbers.PhoneMetadataCollection {
	log.Println("Fetching PhoneNumberMetadata.xml from Github")
	body := fetchURL(metadataURL)

	log.Println("Building new metadata collection")
	collection, err := phonenumbers.BuildPhoneMetadataCollection(body, false, false)
	if err != nil {
		log.Fatalf("Error converting XML: %s", err)
	}

	// now that we've generated our possible patterns we can get rid of possible lengths in our proto buffers
	for _, md := range collection.Metadata {
		md.ClearPossibleLengths()
	}

	// write it out as a protobuf
	data, err := proto.Marshal(collection)
	if err != nil {
		log.Fatalf("Error marshalling metadata: %v", err)
	}

	log.Println("Writing new metadata_bin.go")
	writeFile(metadataPath, generateBinFile("metadataData", data))
	return collection
}

// generates the file contents for a data file
func generateBinFile(variableName string, data []byte) []byte {
	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	w.Write(data)
	w.Close()
	encoded := base64.StdEncoding.EncodeToString(compressed.Bytes())

	// create our output
	output := &bytes.Buffer{}

	// write our header
	output.WriteString("package phonenumbers\n\nvar ")
	output.WriteString(variableName)
	output.WriteString(" = ")
	output.WriteString(strconv.Quote(string(encoded)))
	output.WriteString("\n")
	return output.Bytes()
}

func buildPrefixData(build *prefixBuild) {
	log.Println("Fetching " + build.url + " from Github")
	svnExport(build.dir, build.url)

	// get our top level language directories
	dirs, err := filepath.Glob(fmt.Sprintf("%s/*", build.dir))
	if err != nil {
		log.Fatal(err)
	}

	// for each directory
	languageMappings := make(map[string]map[int]string)
	for _, dir := range dirs {
		// only look at directories
		fi, _ := os.Stat(dir)
		if !fi.IsDir() {
			log.Printf("Ignoring directory: %s\n", dir)
			continue
		}

		// get our language code
		parts := strings.Split(dir, "/")

		// build a map for that directory
		mappings := readMappingsForDir(dir)

		// save it for our language
		languageMappings[parts[1]] = mappings
	}

	output := bytes.Buffer{}
	output.WriteString("package phonenumbers\n\n")
	output.WriteString(fmt.Sprintf("var %s = map[string]string {\n", build.varName))

	for lang, mappings := range languageMappings {
		// iterate through our map, creating our full set of values and prefixes
		prefixes := make([]int, 0, len(mappings))
		seenValues := make(map[string]bool)
		values := make([]string, 0, 255)
		for prefix, value := range mappings {
			prefixes = append(prefixes, prefix)
			_, seen := seenValues[value]
			if !seen {
				values = append(values, value)
				seenValues[value] = true
			}
		}

		// make sure we won't overrun uint16s
		if len(values) > math.MaxUint16 {
			log.Fatal("too many values to represent in uint16")
		}

		// need sorted prefixes for our diff writing to work
		sort.Ints(prefixes)

		// sorted values compress better
		sort.Strings(values)

		// build our reverse mapping from value to offset
		internMappings := make(map[string]uint16)
		for i, value := range values {
			internMappings[value] = uint16(i)
		}

		// write our map
		data := &bytes.Buffer{}

		// first write our values, as length of string and raw bytes
		joinedValues := strings.Join(values, "\n")
		if err = binary.Write(data, binary.LittleEndian, uint32(len(joinedValues))); err != nil {
			log.Fatal(err)
		}
		if err = binary.Write(data, binary.LittleEndian, []byte(joinedValues)); err != nil {
			log.Fatal(err)
		}

		// then then number of prefix / value pairs
		if err = binary.Write(data, binary.LittleEndian, uint32(len(prefixes))); err != nil {
			log.Fatal(err)
		}

		// we write our prefix / value pairs as a varint of the difference of the previous prefix
		// and a uint16 of the value index
		last := 0
		intBuf := make([]byte, 6, 6)
		for _, prefix := range prefixes {
			value := mappings[prefix]
			valueIntern := internMappings[value]
			diff := prefix - last
			l := binary.PutUvarint(intBuf, uint64(diff))
			if err = binary.Write(data, binary.LittleEndian, intBuf[:l]); err != nil {
				log.Fatal(err)
			}
			if err = binary.Write(data, binary.LittleEndian, uint16(valueIntern)); err != nil {
				log.Fatal(err)
			}

			last = prefix
		}

		var compressed bytes.Buffer
		w := gzip.NewWriter(&compressed)
		w.Write(data.Bytes())
		w.Close()
		c := base64.StdEncoding.EncodeToString(compressed.Bytes())
		output.WriteString("\t")
		output.WriteString(strconv.Quote(lang))
		output.WriteString(": ")
		output.WriteString(strconv.Quote(c))
		output.WriteString(",\n")
	}

	output.WriteString("}")
	writeFile(build.srcPath, output.Bytes())
}

func readMappingsForDir(dir string) map[int]string {
	log.Printf("Building map for: %s\n", dir)
	mappings := make(map[int]string)

	files, err := filepath.Glob(dir + "/*.txt")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		body, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		items := strings.Split(file, "/")
		if len(items) != 3 {
			log.Fatalf("file name %s not correct", file)
		}

		for _, line := range strings.Split(string(body), "\n") {
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
			prefix := fields[0]
			prefixInt, err := strconv.Atoi(prefix)
			if err != nil || prefixInt < 0 {
				log.Fatalf("Unable to parse line: %s", line)
			}

			value := strings.TrimSpace(fields[1])
			if value == "" {
				log.Printf("Ignoring empty value: %s", line)
			}

			_, repeat := mappings[prefixInt]
			if repeat {
				log.Fatalf("Repeated prefix for line: %s", line)
			}
			mappings[prefixInt] = fields[1]
		}
	}

	log.Printf("Read %d mappings in %s\n", len(mappings), dir)
	return mappings
}

func main() {
	metadata := buildMetadata()
	buildRegions(metadata)
	buildTimezones()
	buildPrefixData(&carrier)
	buildPrefixData(&geocoding)
}
