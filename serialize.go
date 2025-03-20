package phonenumbers

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

// intStringMap is our data structure for maps from prefixes to a single string
// this is used for our carrier and geocoding maps
type intStringMap struct {
	Map       map[int]string
	MaxLength int
}

func loadPrefixMap(data []byte) (*intStringMap, error) {
	rawBytes, err := decodeUnzip(data)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(rawBytes)

	// ok, first read in our number of values
	var valueSize uint32
	err = binary.Read(reader, binary.LittleEndian, &valueSize)
	if err != nil {
		return nil, err
	}

	// then our values
	valueBytes := make([]byte, valueSize)
	n, err := reader.Read(valueBytes)
	if uint32(n) < valueSize {
		return nil, fmt.Errorf("unable to read all values: %v", err)
	}

	values := strings.Split(string(valueBytes), "\n")

	// read our # of mappings
	var mappingCount uint32
	err = binary.Read(reader, binary.LittleEndian, &mappingCount)
	if err != nil {
		return nil, err
	}

	maxLength := 0
	mappings := make(map[int]string, mappingCount)
	prefix := 0
	for range int(mappingCount) {
		// first read our diff
		diff, err := binary.ReadUvarint(reader)
		if err != nil {
			return nil, err
		}

		prefix += int(diff)

		// then our map
		var valueIntern uint16
		err = binary.Read(reader, binary.LittleEndian, &valueIntern)
		if err != nil || int(valueIntern) >= len(values) {
			return nil, fmt.Errorf("unable to read interned value: %v", err)
		}

		mappings[prefix] = values[valueIntern]

		count := digitCount(prefix)
		if count > maxLength {
			maxLength = count
		}
	}

	// return our values
	return &intStringMap{
		Map:       mappings,
		MaxLength: maxLength,
	}, nil
}

func digitCount(n int) int {
	if n == 0 {
		return 1 // Special case for 0
	}

	if n < 0 {
		n = -n // Handle negative numbers
	}

	return int(math.Floor(math.Log10(float64(n)))) + 1
}

// intStringArrayMap is our map from an int to an array of strings
// this is used for our timezone and region maps
type intStringArrayMap struct {
	Map       map[int][]string
	MaxLength int
}

func loadIntArrayMap(data []byte) (*intStringArrayMap, error) {
	rawBytes, err := decodeUnzip(data)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(rawBytes)

	// ok, first read in our number of values
	var valueSize uint32
	err = binary.Read(reader, binary.LittleEndian, &valueSize)
	if err != nil {
		return nil, err
	}

	// then our values
	valueBytes := make([]byte, valueSize)
	n, err := reader.Read(valueBytes)
	if uint32(n) < valueSize {
		return nil, fmt.Errorf("unable to read all values: %v", err)
	}

	values := strings.Split(string(valueBytes), "\n")

	// read our # of mappings
	var mappingCount uint32
	err = binary.Read(reader, binary.LittleEndian, &mappingCount)
	if err != nil {
		return nil, err
	}

	maxLength := 0
	mappings := make(map[int][]string, mappingCount)
	key := 0
	for range int(mappingCount) {
		// first read our diff
		diff, err := binary.ReadUvarint(reader)
		if err != nil {
			return nil, err
		}

		key += int(diff)

		// then our values
		var valueCount uint8
		if err = binary.Read(reader, binary.LittleEndian, &valueCount); err != nil {
			return nil, err
		}

		keyValues := make([]string, valueCount)
		for i := range int(valueCount) {
			var valueIntern uint16
			err = binary.Read(reader, binary.LittleEndian, &valueIntern)
			if err != nil || int(valueIntern) >= len(values) {
				return nil, fmt.Errorf("unable to read interned value: %v", err)
			}
			keyValues[i] = values[valueIntern]
		}
		mappings[key] = keyValues

		count := digitCount(key)
		if count > maxLength {
			maxLength = count
		}
	}

	// return our values
	return &intStringArrayMap{
		Map:       mappings,
		MaxLength: maxLength,
	}, nil
}

func decodeUnzip(data []byte) ([]byte, error) {
	zipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer zipReader.Close()

	rawBytes, err := io.ReadAll(zipReader)
	if err != nil {
		return nil, err
	}

	return rawBytes, nil
}
