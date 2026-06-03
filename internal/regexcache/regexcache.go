// Package regexcache caches compiled regular expressions keyed by their pattern
// string, so the libphonenumber port avoids recompiling the same
// metadata-derived patterns on every call. It mirrors upstream's
// internal/RegexCache.
package regexcache

import (
	"regexp"
	"sync"
)

var (
	// cache holds frequently used region-specific regular expressions. The
	// initial capacity is set to 100 as this seems to be an optimal value for
	// Android, based on performance measurements.
	cache = make(map[string]*regexp.Regexp)
	mu    sync.RWMutex
)

func read(pattern string) (*regexp.Regexp, bool) {
	mu.RLock()
	v, ok := cache[pattern]
	mu.RUnlock()
	return v, ok
}

func write(pattern string, value *regexp.Regexp) {
	mu.Lock()
	cache[pattern] = value
	mu.Unlock()
}

// For returns the compiled regexp for pattern, compiling and caching it on first
// use. It panics if pattern is not a valid regular expression.
func For(pattern string) *regexp.Regexp {
	regex, found := read(pattern)
	if !found {
		regex = regexp.MustCompile(pattern)
		write(pattern, regex)
	}
	return regex
}
