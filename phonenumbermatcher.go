// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/PhoneNumberMatcher.java.
// Methods are kept in upstream source order to ease syncing.
package phonenumbers

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nyaruka/phonenumbers/v2/internal/regexcache"
	"github.com/nyaruka/phonenumbers/v2/internal/stringbuilder"
)

// limit returns a regular expression quantifier with a lower and upper bound.
func limit(lower, upper int) string {
	if lower < 0 || upper <= 0 || upper < lower {
		panic("invalid limit bounds")
	}
	return "{" + strconv.Itoa(lower) + "," + strconv.Itoa(upper) + "}"
}

var (
	// pubPages matches strings that look like publication pages, e.g. the
	// "211-227 (2003)" in a citation, which is not a telephone number.
	pubPages = regexp.MustCompile(`\d{1,5}-+\d{1,5}\s{0,4}\(\d{1,4}`)

	// slashSeparatedDates matches strings that look like dates using "/" as a
	// separator, e.g. 3/10/2011, 31/10/96 or 08/31/95.
	slashSeparatedDates = regexp.MustCompile(`(?:(?:[0-3]?\d/[01]?\d)|(?:[01]?\d/[0-3]?\d))/(?:[12]\d)?\d{2}`)

	// timeStamps matches timestamps, e.g. "2012-01-02 08:00". The trailing
	// ":\d\d" is covered by timeStampsSuffix.
	timeStamps       = regexp.MustCompile(`[12]\d{3}[-/]?[01]\d[-/]?[0-3]\d +[0-2]\d$`)
	timeStampsSuffix = regexp.MustCompile(`:[0-5]\d`)

	// matchingBrackets checks that brackets match: opening brackets should be
	// closed within a phone number. Anchored for full-string (.matches()).
	matchingBrackets *regexp.Regexp

	// phoneNumberMatcherPattern is the phone number pattern used by find. It is
	// similar to VALID_PHONE_NUMBER but with bounded captures, no surrounding
	// whitespace and no alpha (vanity) digits.
	phoneNumberMatcherPattern *regexp.Regexp

	// leadClassPattern matches (at the start of a candidate) punctuation that may
	// be at the start of a phone number — brackets and plus signs.
	leadClassPattern *regexp.Regexp

	// innerMatches are patterns used to extract phone numbers from a larger
	// phone-number-like pattern, ordered by specificity (white-space last).
	innerMatches = []*regexp.Regexp{
		// Breaks on the slash - e.g. "651-234-2345/332-445-1234"
		regexp.MustCompile(`/+(.*)`),
		// The bracket is inside the capturing group since we consider it part of
		// the phone number. Matches e.g. "(650) 223 3345 (754) 223 3321".
		regexp.MustCompile(`(\([^(]*)`),
		// Breaks on a hyphen with a space on either side - e.g.
		// "12345 - 332-445-1234 is my number."
		regexp.MustCompile(`(?:\p{Z}-|-\p{Z})\p{Z}*(.+)`),
		// Various types of wide hyphens, without enforcing a surrounding space.
		regexp.MustCompile("[‒-―－]\\p{Z}*(.+)"),
		// Breaks on a full stop - e.g. "12345. 332-445-1234 is my number."
		regexp.MustCompile(`\.+\p{Z}*([^.]+)`),
		// Breaks on space - e.g. "3324451234 8002341234"
		regexp.MustCompile(`\p{Z}+(\P{Z}+)`),
	}
)

func init() {
	// Builds the matchingBrackets and phoneNumberMatcherPattern regular
	// expressions. The building blocks below mirror upstream's static block.
	openingParens := "(\\[（［"
	closingParens := ")\\]）］"
	nonParens := "[^" + openingParens + closingParens + "]"

	// Limit on the number of pairs of brackets in a phone number.
	bracketPairLimit := limit(0, 3)
	// An opening bracket at the beginning may not be closed, but subsequent ones
	// should be. The leading bracket may also have been dropped, so a closing
	// bracket may appear first. We limit the sets of brackets to four.
	matchingBrackets = regexp.MustCompile(
		"^(?:[" + openingParens + "])?" + "(?:" + nonParens + "+" + "[" + closingParens + "])?" +
			nonParens + "+" +
			"(?:[" + openingParens + "]" + nonParens + "+[" + closingParens + "])" + bracketPairLimit +
			nonParens + "*$")

	// Limit on the number of leading (plus) characters.
	leadLimit := limit(0, 2)
	// Limit on the number of consecutive punctuation characters.
	punctuationLimit := limit(0, 4)
	// The maximum number of digits allowed in a digit-separated block. As we
	// allow all digits in a single block, set high enough to accommodate the
	// entire national number and the international country code.
	digitBlockLimit := MAX_LENGTH_FOR_NSN + MAX_LENGTH_COUNTRY_CODE
	// Limit on the number of blocks separated by punctuation.
	blockLimit := limit(0, digitBlockLimit)

	// A punctuation sequence allowing white space.
	punctuation := "[" + VALID_PUNCTUATION + "]" + punctuationLimit
	// A digits block without punctuation.
	digitSequence := DIGITS + limit(1, digitBlockLimit)

	leadClassChars := openingParens + PLUS_CHARS
	leadClass := "[" + leadClassChars + "]"
	leadClassPattern = regexp.MustCompile("^(?:" + leadClass + ")")

	// Phone number pattern allowing optional punctuation. REGEX_FLAGS in upstream
	// is UNICODE_CHARACTER_CLASS|CASE_INSENSITIVE; RE2's \p classes are already
	// Unicode-aware, so we only need the (?i) flag (needed by the extension
	// labels in EXTN_PATTERNS_FOR_MATCHING).
	phoneNumberMatcherPattern = regexp.MustCompile(
		"(?i)(?:" + leadClass + punctuation + ")" + leadLimit +
			digitSequence + "(?:" + punctuation + digitSequence + ")" + blockLimit +
			"(?:" + EXTN_PATTERNS_FOR_MATCHING + ")?")
}

// matcherState is the iteration tristate of a phoneNumberMatcher.
type matcherState int

const (
	matcherNotReady matcherState = iota
	matcherReady
	matcherDone
)

// phoneNumberMatcher finds and extracts telephone numbers from text. It is the
// engine behind FindNumbers and is not safe for concurrent use. Vanity numbers
// (using alphabetic digits) are not found.
type phoneNumberMatcher struct {
	// text is the searched text.
	text string
	// preferredRegion is the region to assume for numbers without an
	// international prefix; may be empty / "ZZ".
	preferredRegion string
	// leniency is the degree of validation requested.
	leniency Leniency
	// maxTries is the maximum number of invalid numbers to try before giving up.
	maxTries int

	state       matcherState
	lastMatch   *PhoneNumberMatch
	searchIndex int
}

// newPhoneNumberMatcher creates a matcher over text. country is the region to
// assume for numbers not in international format (empty / "ZZ" if only numbers
// with a leading plus should be considered). maxTries is clamped to >= 0.
func newPhoneNumberMatcher(text, country string, leniency Leniency, maxTries int) *phoneNumberMatcher {
	if maxTries < 0 {
		maxTries = 0
	}
	return &phoneNumberMatcher{
		text:            text,
		preferredRegion: country,
		leniency:        leniency,
		maxTries:        maxTries,
		state:           matcherNotReady,
	}
}

// find attempts to find the next substring at or after index that represents a
// phone number, returning the match or nil if none was found.
func (m *phoneNumberMatcher) find(index int) *PhoneNumberMatch {
	for m.maxTries > 0 {
		loc := phoneNumberMatcherPattern.FindStringIndex(m.text[index:])
		if loc == nil {
			break
		}
		start := index + loc[0]
		candidate := m.text[start : index+loc[1]]

		// Check for extra numbers at the end.
		candidate = trimAfterFirstMatch(SECOND_NUMBER_START_PATTERN, candidate)

		if match := m.extractMatch(candidate, start); match != nil {
			return match
		}

		index = start + len(candidate)
		m.maxTries--
	}
	return nil
}

// trimAfterFirstMatch trims away any characters after the first match of pattern
// in candidate, returning the trimmed version.
func trimAfterFirstMatch(pattern *regexp.Regexp, candidate string) string {
	if loc := pattern.FindStringIndex(candidate); loc != nil {
		return candidate[:loc[0]]
	}
	return candidate
}

// isLatinLetter reports whether a rune is a Latin-script letter. Combining marks
// also return true since we assume they have been added to a preceding Latin
// character.
func isLatinLetter(letter rune) bool {
	// Combining marks are a subset of non-spacing-mark (unicode.Mn).
	if !unicode.IsLetter(letter) && !unicode.Is(unicode.Mn, letter) {
		return false
	}
	// Upstream checks the Unicode block is one of the Latin blocks or the
	// Combining Diacritical Marks block (U+0300..U+036F). RE2/Go has no block
	// tables, so we use the Latin script plus that explicit range.
	return unicode.Is(unicode.Latin, letter) || (letter >= 0x0300 && letter <= 0x036F)
}

func isInvalidPunctuationSymbol(c rune) bool {
	// unicode.Sc == CURRENCY_SYMBOL.
	return c == '%' || unicode.Is(unicode.Sc, c)
}

// extractMatch attempts to extract a match from a candidate substring at offset.
func (m *phoneNumberMatcher) extractMatch(candidate string, offset int) *PhoneNumberMatch {
	// Skip a match that is more likely to be a date.
	if slashSeparatedDates.MatchString(candidate) {
		return nil
	}

	// Skip potential time-stamps.
	if timeStamps.MatchString(candidate) {
		followingText := m.text[offset+len(candidate):]
		if loc := timeStampsSuffix.FindStringIndex(followingText); loc != nil && loc[0] == 0 {
			return nil
		}
	}

	// Try to come up with a valid match given the entire candidate.
	if match := m.parseAndVerify(candidate, offset); match != nil {
		return match
	}

	// If that failed, try to find an "inner match" within this candidate.
	return m.extractInnerMatch(candidate, offset)
}

// extractInnerMatch attempts to extract a match from candidate if the whole
// candidate does not qualify as a match.
func (m *phoneNumberMatcher) extractInnerMatch(candidate string, offset int) *PhoneNumberMatch {
	for _, possibleInnerMatch := range innerMatches {
		isFirstMatch := true
		for _, g := range possibleInnerMatch.FindAllStringSubmatchIndex(candidate, -1) {
			if m.maxTries <= 0 {
				break
			}
			if isFirstMatch {
				// Handle any group before this one too.
				group := trimAfterFirstMatch(UNWANTED_END_CHAR_PATTERN, candidate[:g[0]])
				if match := m.parseAndVerify(group, offset); match != nil {
					return match
				}
				m.maxTries--
				isFirstMatch = false
			}
			group := trimAfterFirstMatch(UNWANTED_END_CHAR_PATTERN, candidate[g[2]:g[3]])
			if match := m.parseAndVerify(group, offset+g[2]); match != nil {
				return match
			}
			m.maxTries--
		}
	}
	return nil
}

// parseAndVerify parses a phone number from candidate and verifies it matches
// the requested leniency, returning a PhoneNumberMatch or nil.
func (m *phoneNumberMatcher) parseAndVerify(candidate string, offset int) *PhoneNumberMatch {
	// Check the candidate doesn't contain any formatting which would indicate
	// that it really isn't a phone number.
	if !matchingBrackets.MatchString(candidate) || pubPages.MatchString(candidate) {
		return nil
	}

	// If leniency is VALID or stricter, skip numbers surrounded by Latin
	// alphabetic characters, to skip cases like abc8005001234 or 8005001234def.
	if m.leniency >= VALID {
		// If the candidate is not at the start of the text, and does not start
		// with phone-number punctuation, check the previous character.
		if offset > 0 && leadClassPattern.FindStringIndex(candidate) == nil {
			previousChar, _ := utf8.DecodeLastRuneInString(m.text[:offset])
			if isInvalidPunctuationSymbol(previousChar) || isLatinLetter(previousChar) {
				return nil
			}
		}
		lastCharIndex := offset + len(candidate)
		if lastCharIndex < len(m.text) {
			nextChar, _ := utf8.DecodeRuneInString(m.text[lastCharIndex:])
			if isInvalidPunctuationSymbol(nextChar) || isLatinLetter(nextChar) {
				return nil
			}
		}
	}

	number := &PhoneNumber{}
	if err := ParseAndKeepRawInputToNumber(candidate, m.preferredRegion, number); err != nil {
		// ignore and continue
		return nil
	}

	if m.leniency.Verify(number, candidate) {
		// We used ParseAndKeepRawInput to create this number, but for now we
		// don't return the extra values parsed.
		number.CountryCodeSource = nil
		number.RawInput = nil
		number.PreferredDomesticCarrierCode = nil
		return newPhoneNumberMatch(offset, candidate, number)
	}
	return nil
}

// getNationalNumberGroups returns the national-number part of a number,
// formatted without any national prefix, as the set of digit blocks that would
// be formatted together following standard formatting rules.
func getNationalNumberGroups(number *PhoneNumber) []string {
	// This will be in the format +CC-DG1-DG2-DGX;ext=EXT.
	rfc3966Format := Format(number, RFC3966)
	// Remove the extension part before splitting into groups.
	endIndex := strings.IndexByte(rfc3966Format, ';')
	if endIndex < 0 {
		endIndex = len(rfc3966Format)
	}
	// The country code is followed by a '-'.
	startIndex := strings.IndexByte(rfc3966Format, '-') + 1
	return strings.Split(rfc3966Format[startIndex:endIndex], "-")
}

// getNationalNumberGroupsWithPattern returns the national-number digit blocks
// formatted according to the formatting pattern passed in.
func getNationalNumberGroupsWithPattern(number *PhoneNumber, formattingPattern *NumberFormat) []string {
	nationalSignificantNumber := GetNationalSignificantNumber(number)
	return strings.Split(formatNsnUsingPattern(nationalSignificantNumber, formattingPattern, RFC3966), "-")
}

// CheckNumberGroupingIsValid checks the digit groups of candidate against the
// expected grouping for number using checker, falling back to any alternate
// formats for the number's country calling code.
func CheckNumberGroupingIsValid(
	number *PhoneNumber,
	candidate string,
	checker func(number *PhoneNumber, normalizedCandidate string, expectedNumberGroups []string) bool) bool {

	normalizedCandidate := normalizeDigits(candidate, true /* keep non-digits */)
	formattedNumberGroups := getNationalNumberGroups(number)
	if checker(number, normalizedCandidate, formattedNumberGroups) {
		return true
	}
	// If this didn't pass, see if any alternate formats match instead.
	alternateFormats := getAlternateFormatsForCountryCallingCode(number.GetCountryCode())
	nationalSignificantNumber := GetNationalSignificantNumber(number)
	if alternateFormats != nil {
		for _, alternateFormat := range alternateFormats.GetNumberFormat() {
			if len(alternateFormat.GetLeadingDigitsPattern()) > 0 {
				// There is only one leading digits pattern for alternate formats.
				pattern := regexcache.For(alternateFormat.GetLeadingDigitsPattern()[0])
				if loc := pattern.FindStringIndex(nationalSignificantNumber); loc == nil || loc[0] != 0 {
					// Leading digits don't match (lookingAt); try another one.
					continue
				}
			}
			formattedNumberGroups = getNationalNumberGroupsWithPattern(number, alternateFormat)
			if checker(number, normalizedCandidate, formattedNumberGroups) {
				return true
			}
		}
	}
	return false
}

// AllNumberGroupsRemainGrouped reports that the number groups found in the
// candidate are not broken up differently from how the number would be
// formatted (the STRICT_GROUPING check).
func AllNumberGroupsRemainGrouped(
	number *PhoneNumber,
	normalizedCandidate string,
	formattedNumberGroups []string) bool {

	fromIndex := 0
	if number.GetCountryCodeSource() != PhoneNumber_FROM_DEFAULT_COUNTRY {
		// First skip the country code if the normalized candidate contained it.
		countryCode := strconv.Itoa(int(number.GetCountryCode()))
		fromIndex = strings.Index(normalizedCandidate, countryCode) + len(countryCode)
	}
	// Check each group of consecutive digits is not broken into separate
	// groupings in the normalizedCandidate string.
	for i := 0; i < len(formattedNumberGroups); i++ {
		// Fails if the substring of normalizedCandidate starting from fromIndex
		// doesn't contain the consecutive digits in formattedNumberGroups[i].
		idx := strings.Index(normalizedCandidate[fromIndex:], formattedNumberGroups[i])
		if idx < 0 {
			return false
		}
		// Move fromIndex forward.
		fromIndex = fromIndex + idx + len(formattedNumberGroups[i])
		if i == 0 && fromIndex < len(normalizedCandidate) {
			// We are at the position right after the NDC. We get the region used
			// for formatting based on the country code rather than the number
			// itself, as we don't need to distinguish between different countries
			// with the same country calling code and this is faster.
			region := GetRegionCodeForCountryCode(int(number.GetCountryCode()))
			nextChar, _ := utf8.DecodeRuneInString(normalizedCandidate[fromIndex:])
			if GetNddPrefixForRegion(region, true) != "" && unicode.IsDigit(nextChar) {
				// There is no formatting symbol after the NDC. In this case we
				// only accept the number if there is no formatting symbol at all
				// in the number, except for extensions. Only important for
				// countries with national prefixes.
				nationalSignificantNumber := GetNationalSignificantNumber(number)
				return strings.HasPrefix(
					normalizedCandidate[fromIndex-len(formattedNumberGroups[i]):], nationalSignificantNumber)
			}
		}
	}
	// Make sure we haven't mistakenly used the extension to match the last group
	// of the subscriber number. The extension cannot have formatting in between.
	return strings.Contains(normalizedCandidate[fromIndex:], number.GetExtension())
}

// AllNumberGroupsAreExactlyPresent reports that the groups of digits in the
// candidate exactly match how the number would be formatted (the EXACT_GROUPING
// check).
func AllNumberGroupsAreExactlyPresent(
	number *PhoneNumber,
	normalizedCandidate string,
	formattedNumberGroups []string) bool {

	candidateGroups := splitNonDigits(normalizedCandidate)
	// Set this to the last group, skipping it if the number has an extension.
	candidateNumberGroupIndex := len(candidateGroups) - 1
	if number.GetExtension() != "" {
		candidateNumberGroupIndex = len(candidateGroups) - 2
	}

	// First check if the national significant number is formatted as a block.
	// We use contains and not equals, since the national significant number may
	// be present with a prefix such as a national number prefix, or the country
	// code itself.
	if len(candidateGroups) == 1 ||
		strings.Contains(candidateGroups[candidateNumberGroupIndex], GetNationalSignificantNumber(number)) {
		return true
	}
	// Starting from the end, go through in reverse, excluding the first group,
	// and check the candidate and number groups are the same.
	for formattedNumberGroupIndex := len(formattedNumberGroups) - 1; formattedNumberGroupIndex > 0 && candidateNumberGroupIndex >= 0; formattedNumberGroupIndex-- {
		if candidateGroups[candidateNumberGroupIndex] != formattedNumberGroups[formattedNumberGroupIndex] {
			return false
		}
		candidateNumberGroupIndex--
	}
	// Now check the first group. There may be a national prefix at the start, so
	// we only check that the candidate group ends with the formatted number group.
	return candidateNumberGroupIndex >= 0 &&
		strings.HasSuffix(candidateGroups[candidateNumberGroupIndex], formattedNumberGroups[0])
}

// splitNonDigits splits s on runs of non-digits, replicating Java's
// String.split(\D+) semantics (leading empty strings are kept, trailing empty
// strings are dropped).
func splitNonDigits(s string) []string {
	groups := NON_DIGITS_PATTERN.Split(s, -1)
	// Java's String.split with the default limit drops trailing empty strings.
	for len(groups) > 0 && groups[len(groups)-1] == "" {
		groups = groups[:len(groups)-1]
	}
	if len(groups) == 0 {
		groups = []string{""}
	}
	return groups
}

// ContainsMoreThanOneSlashInNationalNumber reports whether candidate contains
// more than one slash in the national-number portion (used to reject dates).
func ContainsMoreThanOneSlashInNationalNumber(number *PhoneNumber, candidate string) bool {
	firstSlashInBodyIndex := strings.IndexByte(candidate, '/')
	if firstSlashInBodyIndex < 0 {
		// No slashes, this is okay.
		return false
	}
	// Now look for a second one.
	secondSlashInBodyIndex := strings.IndexByte(candidate[firstSlashInBodyIndex+1:], '/')
	if secondSlashInBodyIndex < 0 {
		// Only one slash, this is okay.
		return false
	}
	secondSlashInBodyIndex += firstSlashInBodyIndex + 1

	// If the first slash is after the country calling code, this is permitted.
	candidateHasCountryCode := number.GetCountryCodeSource() == PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN ||
		number.GetCountryCodeSource() == PhoneNumber_FROM_NUMBER_WITHOUT_PLUS_SIGN
	if candidateHasCountryCode &&
		NormalizeDigitsOnly(candidate[:firstSlashInBodyIndex]) == strconv.Itoa(int(number.GetCountryCode())) {
		// Any more slashes and this is illegal.
		return strings.Contains(candidate[secondSlashInBodyIndex+1:], "/")
	}
	return true
}

// ContainsOnlyValidXChars reports whether every 'x'/'X' in candidate represents
// a carrier code or an extension sign (and not, say, a vanity-number letter).
func ContainsOnlyValidXChars(number *PhoneNumber, candidate string) bool {
	// The characters 'x' and 'X' can be (1) a carrier code, in which case they
	// always precede the national significant number or (2) an extension sign,
	// in which case they always precede the extension number. We assume a
	// carrier code is more than 1 digit, so the first case has to have more than
	// 1 consecutive 'x' or 'X', whereas the second can only have exactly 1. We
	// ignore the character if it appears as the last character of the string.
	for index := 0; index < len(candidate)-1; index++ {
		charAtIndex := candidate[index]
		if charAtIndex == 'x' || charAtIndex == 'X' {
			charAtNextIndex := candidate[index+1]
			if charAtNextIndex == 'x' || charAtNextIndex == 'X' {
				// Carrier code case: the 'X's always precede the national
				// significant number.
				index++
				if IsNumberMatchWithOneNumber(number, candidate[index:]) != NSN_MATCH {
					return false
				}
				// Extension sign case: the 'x' or 'X' should always precede the
				// extension number.
			} else if NormalizeDigitsOnly(candidate[index:]) != number.GetExtension() {
				return false
			}
		}
	}
	return true
}

// IsNationalPrefixPresentIfRequired reports whether, if a national prefix is
// required to format number, it was present in the raw input.
func IsNationalPrefixPresentIfRequired(number *PhoneNumber) bool {
	// First, check how we deduced the country code. If it was written in
	// international format, then the national prefix is not required.
	if number.GetCountryCodeSource() != PhoneNumber_FROM_DEFAULT_COUNTRY {
		return true
	}
	phoneNumberRegion := GetRegionCodeForCountryCode(int(number.GetCountryCode()))
	metadata := getMetadataForRegion(phoneNumberRegion)
	if metadata == nil {
		return true
	}
	// Check if a national prefix should be present when formatting this number.
	nationalNumber := GetNationalSignificantNumber(number)
	formatRule := chooseFormattingPatternForNumber(metadata.GetNumberFormat(), nationalNumber)
	// To do this, we check that a national prefix formatting rule was present and
	// that it wasn't just the first-group symbol ($1) with punctuation.
	if formatRule != nil && len(formatRule.GetNationalPrefixFormattingRule()) > 0 {
		if formatRule.GetNationalPrefixOptionalWhenFormatting() {
			// The national-prefix is optional in these cases, so we don't need to
			// check if it was present.
			return true
		}
		if formattingRuleHasFirstGroupOnly(formatRule.GetNationalPrefixFormattingRule()) {
			// National Prefix not needed for this number.
			return true
		}
		// Normalize the remainder.
		rawInputCopy := NormalizeDigitsOnly(number.GetRawInput())
		rawInput := stringbuilder.NewString(rawInputCopy)
		// Check if we found a national prefix and/or carrier code at the start of
		// the raw input, and return the result.
		return maybeStripNationalPrefixAndCarrierCode(rawInput, metadata, nil)
	}
	return true
}

// hasNext reports whether there is another match, finding it as a side effect.
func (m *phoneNumberMatcher) hasNext() bool {
	if m.state == matcherNotReady {
		m.lastMatch = m.find(m.searchIndex)
		if m.lastMatch == nil {
			m.state = matcherDone
		} else {
			m.searchIndex = m.lastMatch.End()
			m.state = matcherReady
		}
	}
	return m.state == matcherReady
}

// next returns the next match, or nil if there is none (callers should guard
// with hasNext).
func (m *phoneNumberMatcher) next() *PhoneNumberMatch {
	if !m.hasNext() {
		return nil
	}
	result := m.lastMatch
	m.lastMatch = nil
	m.state = matcherNotReady
	return result
}
