// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/internal/RegexBasedMatcher.java + .../internal/MatcherApi.java.

// Package regexbasedmatcher matches national numbers against the patterns in
// phone-number metadata.
package regexbasedmatcher

import (
	"regexp"

	"github.com/nyaruka/phonenumbers/v2/internal/regexcache"
	"github.com/nyaruka/phonenumbers/v2/metadata"
)

// MatchNationalNumber reports whether number (a string of decimal digits)
// matches the national number pattern in numberDesc.
func MatchNationalNumber(number string, numberDesc *metadata.PhoneNumberDesc, allowPrefixMatch bool) bool {
	nationalNumberPattern := numberDesc.GetNationalNumberPattern()
	// We don't want to consider it a prefix match when matching non-empty input
	// against an empty pattern.
	if len(nationalNumberPattern) == 0 {
		return false
	}
	return match(number, regexcache.For(nationalNumberPattern), allowPrefixMatch)
}

func match(number string, pattern *regexp.Regexp, allowPrefixMatch bool) bool {
	ind := pattern.FindStringIndex(number)
	if len(ind) == 0 || ind[0] != 0 {
		return false
	}
	patP := `^(?:` + pattern.String() + `)$` // Strictly match
	pat := regexcache.For(patP)
	return pat.MatchString(number) || allowPrefixMatch
}
