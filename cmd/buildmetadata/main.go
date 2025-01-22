package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Accompany-Health/phonenumbers"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := buildMetadata(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildMetadata() error {
	fmt.Print("Cloning upstream repo... ")

	if err := cloneUpstreamRepo("https://github.com/google/libphonenumber.git"); err != nil {
		return err
	}

	fmt.Print("OK\nBuilding number metadata...")

	metadata, err := buildNumberMetadata("resources/PhoneNumberMetadata.xml", "NumberData", "metadata_bin.go", false)
	if err != nil {
		return err
	}

	fmt.Print("OK\nBuilding short number metadata...")

	_, err = buildNumberMetadata("resources/ShortNumberMetadata.xml", "ShortNumberData", "shortnumber_metadata_bin.go", true)
	if err != nil {
		return err
	}

	fmt.Print("OK\nBuilding region metadata...")

	if err := buildRegionMetadata(metadata, "RegionData", "countrycode_to_region_bin.go"); err != nil {
		return err
	}

	fmt.Print("OK\nBuilding timezone metadata...")

	if err := buildTimezoneMetadata("resources/timezones/map_data.txt", "TimezoneData", "prefix_to_timezone_bin.go"); err != nil {
		return err
	}

	fmt.Println("OK\nBuilding carrier prefix metadata...")

	if err := buildPrefixMetadata("resources/carrier", "CarrierData", "prefix_to_carriers_bin.go"); err != nil {
		return err
	}

	fmt.Println("Building geographic prefix metadata...")

	if err := buildPrefixMetadata("resources/geocoding", "GeocodingData", "prefix_to_geocodings_bin.go"); err != nil {
		return err
	}

	return nil
}

func cloneUpstreamRepo(url string) error {
	os.RemoveAll("_build")

	cmd := exec.Command("git", "clone", "--depth=1", url, "_build")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning upstream repo: %w", err)
	}

	return nil
}

func buildNumberMetadata(srcFile, varName, dstFile string, short bool) (*phonenumbers.PhoneMetadataCollection, error) {
	body, err := os.ReadFile("_build/" + srcFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", srcFile, err)
	}

	collection, err := phonenumbers.BuildPhoneMetadataCollection(body, false, false, short)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", srcFile, err)
	}

	data, err := proto.Marshal(collection)
	if err != nil {
		return nil, fmt.Errorf("error marshaling metadata as protobuf: %w", err)
	}

	if err := os.WriteFile("gen/"+dstFile, generateBinFile(varName, data), os.FileMode(0664)); err != nil {
		return nil, fmt.Errorf("error writing %s: %w", dstFile, err)
	}

	return collection, nil
}

func buildRegionMetadata(metadata *phonenumbers.PhoneMetadataCollection, varName, dstFile string) error {
	regionMap := phonenumbers.BuildCountryCodeToRegionMap(metadata)

	// generate our map data
	data, err := renderMap(regionMap)
	if err != nil {
		return fmt.Errorf("error generating %s: %w", dstFile, err)
	}

	if err := os.WriteFile("gen/"+dstFile, generateBinFile(varName, data), os.FileMode(0664)); err != nil {
		return fmt.Errorf("error writing %s: %w", dstFile, err)
	}

	return nil
}

func buildTimezoneMetadata(srcFile, varName, dstFile string) error {
	body, err := os.ReadFile("_build/" + srcFile)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", srcFile, err)
	}

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
			return fmt.Errorf("invalid format in timezone file: %s", line)
		}

		zones := strings.Split(fields[1], "&")
		if len(zones) < 1 {
			return fmt.Errorf("invalid format in timezone file: %s", line)
		}

		// parse our prefix
		prefix, err := strconv.Atoi(fields[0])
		if err != nil {
			return fmt.Errorf("invalid prefix in line: %s", line)
		}
		prefixMap[prefix] = zones
	}

	// generate our map data
	data, err := renderMap(prefixMap)
	if err != nil {
		return fmt.Errorf("error generating %s: %w", dstFile, err)
	}

	if err := os.WriteFile("gen/"+dstFile, generateBinFile(varName, data), os.FileMode(0664)); err != nil {
		return fmt.Errorf("error writing %s: %w", dstFile, err)
	}

	return nil
}

func buildPrefixMetadata(srcDir, varName, dstFile string) error {
	// get our top level language directories
	dirs, err := filepath.Glob(fmt.Sprintf("_build/%s/*", srcDir))
	if err != nil {
		return err
	}

	// for each directory
	languageMappings := make(map[string]map[int]string)
	for _, dir := range dirs {
		fi, _ := os.Stat(dir) // only look at directories
		if !fi.IsDir() {
			continue
		}

		// build a map for that directory
		mappings, err := readMappingsForDir(dir)
		if err != nil {
			return fmt.Errorf("error reading mappings for %s: %w", dir, err)
		}

		// save it for our language
		languageMappings[filepath.Base(dir)] = mappings
	}

	output := bytes.Buffer{}
	output.WriteString("package gen\n\n")
	output.WriteString(fmt.Sprintf("var %s = map[string]string {\n", varName))

	langs := maps.Keys(languageMappings)
	sort.Strings(langs)

	for _, lang := range langs {
		mappings := languageMappings[lang]

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
			return fmt.Errorf("too many values to represent in uint16")
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
			return err
		}
		if err = binary.Write(data, binary.LittleEndian, []byte(joinedValues)); err != nil {
			return err
		}

		// then then number of prefix / value pairs
		if err = binary.Write(data, binary.LittleEndian, uint32(len(prefixes))); err != nil {
			return err
		}

		// we write our prefix / value pairs as a varint of the difference of the previous prefix
		// and a uint16 of the value index
		last := 0
		intBuf := make([]byte, 6)
		for _, prefix := range prefixes {
			value := mappings[prefix]
			valueIntern := internMappings[value]
			diff := prefix - last
			l := binary.PutUvarint(intBuf, uint64(diff))
			if err = binary.Write(data, binary.LittleEndian, intBuf[:l]); err != nil {
				return err
			}
			if err = binary.Write(data, binary.LittleEndian, uint16(valueIntern)); err != nil {
				return err
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

	if err := os.WriteFile("gen/"+dstFile, output.Bytes(), os.FileMode(0664)); err != nil {
		return fmt.Errorf("error writing %s: %w", dstFile, err)
	}

	return nil
}

func renderMap(prefixMap map[int][]string) ([]byte, error) {
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
		return nil, err
	}
	if err := binary.Write(data, binary.LittleEndian, []byte(joinedValues)); err != nil {
		return nil, err
	}

	// then the number of keys
	if err := binary.Write(data, binary.LittleEndian, uint32(len(keys))); err != nil {
		return nil, err
	}

	// we write our key / value pairs as a varint of the difference of the previous prefix
	// and a uint16 of the value index
	last := 0
	intBuf := make([]byte, 6)
	for _, key := range keys {
		// first write our prefix
		diff := key - last
		l := binary.PutUvarint(intBuf, uint64(diff))
		if err := binary.Write(data, binary.LittleEndian, intBuf[:l]); err != nil {
			return nil, err
		}

		// then our values
		values := prefixMap[key]

		// write our number of values
		if err := binary.Write(data, binary.LittleEndian, uint8(len(values))); err != nil {
			return nil, err
		}

		// then each value as the interned index
		for _, v := range values {
			valueIntern := internMap[v]
			if err := binary.Write(data, binary.LittleEndian, uint16(valueIntern)); err != nil {
				return nil, err
			}
		}

		last = key
	}

	return data.Bytes(), nil
}

// generates the file contents for a data file
func generateBinFile(varName string, data []byte) []byte {
	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	w.Write(data)
	w.Close()
	encoded := base64.StdEncoding.EncodeToString(compressed.Bytes())

	// create our output
	output := &bytes.Buffer{}

	// write our header
	output.WriteString("package gen\n\nvar ")
	output.WriteString(varName)
	output.WriteString(" = ")
	output.WriteString(strconv.Quote(string(encoded)))
	output.WriteString("\n")
	return output.Bytes()
}

func readMappingsForDir(dir string) (map[int]string, error) {
	lang := filepath.Base(dir)
	mappings := make(map[int]string)

	files, err := filepath.Glob(dir + "/*.txt")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			return nil, err
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
				return nil, fmt.Errorf("error parsing line: %s", line)
			}

			value := strings.TrimSpace(fields[1])
			if value == "" {
				continue
			}

			_, repeat := mappings[prefixInt]
			if repeat {
				return nil, fmt.Errorf("found repeated prefix on line: %s", line)
			}
			mappings[prefixInt] = fields[1]
		}
	}

	fmt.Printf(" > read %d mappings for %s\n", len(mappings), lang)

	return mappings, nil
}
