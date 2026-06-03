// Port of internal/prefixmapper/* (shared prefix lookup) from google/libphonenumber.
//
// A Mapper resolves the value (carrier name, geocoding description, …) mapped to
// the longest matching numeric prefix of an E164 number. It is shared by the
// carrier and geocoding packages, each of which constructs one over its own
// embedded per-language data.
package prefixmapper

import (
	"embed"
	"errors"
	"io/fs"
	"strconv"
	"sync"

	"github.com/nyaruka/phonenumbers/v2/internal/serialize"
)

// Mapper loads "<dir>/<language>.txt.gz" prefix maps from an embedded
// filesystem on demand, caching them per language. It is safe for concurrent
// use.
type Mapper struct {
	mu      sync.Mutex
	langMap map[string]*serialize.IntStringMap
	dataFS  embed.FS
	dir     string
}

// New returns a Mapper backed by dataFS, reading language files from dir.
func New(dataFS embed.FS, dir string) *Mapper {
	return &Mapper{
		langMap: make(map[string]*serialize.IntStringMap),
		dataFS:  dataFS,
		dir:     dir,
	}
}

func (m *Mapper) lazyLoadPrefixes(language string) (*serialize.IntStringMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if we already have prefixes (or nil if they don't exist) return that
	prefixes, ok := m.langMap[language]
	if ok {
		return prefixes, nil
	}

	// try to load the data file for this language
	data, err := m.dataFS.ReadFile(m.dir + "/" + language + ".txt.gz")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	if data != nil {
		prefixes, err = serialize.LoadPrefixMap(data)
		if err != nil {
			return nil, err
		}
		m.langMap[language] = prefixes
	} else {
		m.langMap[language] = nil // language doesn't have data
	}

	return m.langMap[language], nil
}

// ValueForNumber returns the value mapped to the longest matching prefix of the
// given E164 string (as produced by Format with E164, leading + included),
// along with the matched prefix, trying prefixes from maxLength digits downward.
// It returns "" and 0 if no prefix matches or the language has no data.
func (m *Mapper) ValueForNumber(language string, maxLength int, e164 string) (string, int, error) {
	prefixes, err := m.lazyLoadPrefixes(language)
	if err != nil || prefixes == nil {
		return "", 0, err
	}

	l := len(e164)
	if maxLength > l {
		maxLength = l
	}
	// e164[0] is the leading '+', which strconv.Atoi reads as a sign, so e.g.
	// "+86137" parses to 86137 — the prefixes are keyed without the plus.
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
