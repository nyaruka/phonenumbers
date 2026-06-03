package regexcache

import (
	"regexp"
	"testing"
)

func TestRegexCacheWrite(t *testing.T) {
	pattern1 := "TestRegexCacheWrite"
	if _, found1 := read(pattern1); found1 {
		t.Errorf("pattern |%v| is in the cache", pattern1)
	}
	regex1 := For(pattern1)
	cachedRegex1, found1 := read(pattern1)
	if !found1 {
		t.Errorf("pattern |%v| is not in the cache", pattern1)
	}
	if regex1 != cachedRegex1 {
		t.Error("expected the same instance, but got a different one")
	}
	pattern2 := pattern1 + "."
	if _, found2 := read(pattern2); found2 {
		t.Errorf("pattern |%v| is in the cache", pattern2)
	}
}

func TestRegexCacheRead(t *testing.T) {
	pattern1 := "TestRegexCacheRead"
	if _, found1 := read(pattern1); found1 {
		t.Errorf("pattern |%v| is in the cache", pattern1)
	}
	regex1 := regexp.MustCompile(pattern1)
	write(pattern1, regex1)
	if cachedRegex1 := For(pattern1); cachedRegex1 != regex1 {
		t.Error("expected the same instance, but got a different one")
	}
	cachedRegex1, found1 := read(pattern1)
	if !found1 {
		t.Errorf("pattern |%v| is not in the cache", pattern1)
	}
	if cachedRegex1 != regex1 {
		t.Error("expected the same instance, but got a different one")
	}
	pattern2 := pattern1 + "."
	if _, found2 := read(pattern2); found2 {
		t.Errorf("pattern |%v| is in the cache", pattern2)
	}
}
