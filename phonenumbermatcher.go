package phonenumbers

import (
	"io"
	"regexp"
	"strconv"
	"unicode"
)

// A stateful class that finds and extracts telephone numbers fom text.
//
// Vanity numbers (phone numbers using alphabetic digits such as '1-800-SIX-FLAGS' are not found.
type PhoneNumberMatcher struct {
	text            string
	preferredRegion string
	leniency        Leniency
	maxTries        int
	state           int
	lastMatch       *PhoneNumberMatch
	searchIndex     int
}

const (
	notReady = 0
	ready    = 1
	done     = 2
)

var (
	// The phone number pattern used by {@link #find}, similar to
	// {@code PhoneNumberUtil.VALID_PHONE_NUMBER}, but with the following differences:
	// <ul>
	//   <li>All captures are limited in order to place an upper bound to the text matched by the
	//       pattern.
	// <ul>
	//   <li>Leading punctuation / plus signs are limited.
	//   <li>Consecutive occurrences of punctuation are limited.
	//   <li>Number of digits is limited.
	// </ul>
	//   <li>No whitespace is allowed at the start or end.
	//   <li>No alpha digits (vanity numbers such as 1-800-SIX-FLAGS) are currently supported.
	// </ul>
	PATTERN = regexp.MustCompile("(?i)(?:\\+){0,1}(?:" + LEAD_CLASS + PUNCTUATION + ")" + LEAD_LIMIT + DIGIT_SEQUENCE + "(?:" + PUNCTUATION + DIGIT_SEQUENCE + ")" + BLOCK_LIMIT + "(?:" + EXTN_PATTERNS_FOR_MATCHING + ")?")

	//  Matches strings that look like publication pages. Example:
	//  <pre>Computing Complete Answers to Queries in the Presence of Limited Access Patterns.
	//  Chen Li. VLDB J. 12(3): 211-227 (2003).</pre>
	//
	//  The string "211-227 (2003)" is not a telephone number.
	PUB_PAGES = regexp.MustCompile("\\d{1,5}-+\\d{1,5}\\s{0,4}\\(\\d{1,4}")

	// Matches strings that look like dates using "/" as a separator. Examples: 3/10/2011, 31/10/96 or 08/31/95.
	SLASH_SEPARATED_DATES = regexp.MustCompile("(?:(?:[0-3]?\\d/[01]?\\d)|(?:[01]?\\d/[0-3]?\\d))/(?:[12]\\d)?\\d{2}")

	//  Matches timestamps. Examples: "2012-01-02 08:00". Note that the reg-ex does not include the trailing ":\d\d" -- that is covered by TIME_STAMPS_SUFFIX.
	TIME_STAMPS        = regexp.MustCompile("[12]\\d{3}[-/]?[01]\\d[-/]?[0-3]\\d +[0-2]\\d$")
	TIME_STAMPS_SUFFIX = regexp.MustCompile(":[0-5]\\d")

	// Pattern to check that brackets match. Opening brackets should be closed within a phone number.
	// This also checks that there is something inside the brackets. Having no brackets at all is also
	// fine.
	// An opening bracket at the beginning may not be closed, but subsequent ones should be.  It's
	// also possible that the leading bracket was dropped, so we shouldn't be surprised if we see a
	// closing bracket first. We limit the sets of brackets in a phone number to four.
	MATCHING_BRACKETS = regexp.MustCompile("(?:[" + OPENING_PARENS + "])?" + "(?:" + NON_PARENS + "+" + "[" + CLOSING_PARENS + "])?" + NON_PARENS + "+" + "(?:[" + OPENING_PARENS + "]" + NON_PARENS + "+[" + CLOSING_PARENS + "])" + BRACKET_PAIR_LIMIT + NON_PARENS + "*")

	// Patterns used to extract phone numbers from a larger phone-number-like pattern. These are
	// ordered according to specificity. For example, white-space is last since that is frequently
	// used in numbers, not just to separate two numbers. We have separate patterns since we don't
	// want to break up the phone-number-like text on more than one different kind of symbol at one
	// time, although symbols of the same type (e.g. space) can be safely grouped together.
	//
	// Note that if there is a match, we will always check any text found up to the first match as
	// well.
	INNER_MATCHES = []*regexp.Regexp{
		// Breaks on the slash - e.g. "651-234-2345/332-445-1234"
		regexp.MustCompile("/+(.*)"),
		// Note that the bracket here is inside the capturing group, since we consider it part of the
		// phone number. Will match a pattern like "(650) 223 3345 (754) 223 3321".
		regexp.MustCompile("(\\([^(]*)"),
		// Breaks on a hyphen - e.g. "12345 - 332-445-1234 is my number."
		// We require a space on either side of the hyphen for it to be considered a separator.
		regexp.MustCompile("(?:\\p{Z}-|-\\p{Z})\\p{Z}*(.+)"),
		// Various types of wide hyphens. Note we have decided not to enforce a space here, since it's
		// possible that it's supposed to be used to break two numbers without spaces, and we haven't
		// seen many instances of it used within a number.
		regexp.MustCompile("[\u2012-\u2015\uFF0D]\\p{Z}*(.+)"),
		// Breaks on a full stop - e.g. "12345. 332-445-1234 is my number."
		regexp.MustCompile("\\.+\\p{Z}*([^.]+)"),
		// Breaks on space - e.g. "3324451234 8002341234"
		regexp.MustCompile("\\p{Z}+(\\P{Z}+)"),
	}

	//  Punctuation that may be at the start of a phone number - brackets and plus signs.
	LEAD_CLASS   = OPENING_PARENS + PLUS_CHARS
	LEAD_PATTERN = regexp.MustCompile(LEAD_CLASS)

	// Builds the MATCHING_BRACKETS and PATTERN regular expressions. The building blocks below exist to make the pattern more easily understood.
	OPENING_PARENS = "\\(\\[\uFF08\uFF3B"
	CLOSING_PARENS = "\\)\\]\uFF09\uFF3D"
	NON_PARENS     = "[^" + OPENING_PARENS + CLOSING_PARENS + "]"

	// Limit on the number of pairs of brackets in a phone number.
	BRACKET_PAIR_LIMIT = "{0,3}"

	// Limit on the number of leading (plus) characters.
	LEAD_LIMIT = "{0,2}"

	// Limit on the number of consecutive punctuation characters.
	PUNCTIATION_LIMIT = "{0,4}"

	// The maximum number of digits allowed in a digit-separated block. As we allow all digits in a
	//single block, set high enough to accommodate the entire national number and the international
	//country code.
	DIGIT_BLOCK_LIMIT = 17 + 3

	// Limit on the number of blocks separated by punctuation. Uses digitBlockLimit since some
	// formats use spaces to separate each digit.
	BLOCK_LIMIT = "{0," + strconv.Itoa(DIGIT_BLOCK_LIMIT) + "}"

	// A punctuation sequence allowing white space.
	PUNCTUATION = "[" + VALID_PUNCTUATION + "]" + PUNCTIATION_LIMIT

	// A digits block without punctuation.
	DIGIT_SEQUENCE = "\\d{1," + strconv.Itoa(DIGIT_BLOCK_LIMIT) + "}"
)

// Creates a new instance.
//
// Arguments:
// text -- The character sequence that we will search
// country -- The country to assume for phone numbers not written in
//            international format (with a leading plus, or with the
//            international dialing prefix of the specified region). May be
//            "ZZ" if only numbers with a leading plus should be considered.
func NewPhoneNumberMatcher(text string, region string) PhoneNumberMatcher {
	m := PhoneNumberMatcher{
		text:            text,
		preferredRegion: region,
		leniency:        Leniency(1),
		maxTries:        65535,
		state:           notReady,
		lastMatch:       nil,
		searchIndex:     0,
	}

	return m
}

// Trims away any characters after the first match of pattern in
// candidate, returning the trimmed version.
func (*PhoneNumberMatcher) trimAfterFirstMatch(pattern *regexp.Regexp, candidate string) string {
	trailingCharsMatch := pattern.FindStringIndex(candidate)
	if trailingCharsMatch != nil {
		candidate = candidate[:trailingCharsMatch[0]]
	}
	return candidate
}

func (*PhoneNumberMatcher) isInvalidPunctuationSymbol(char rune) bool {
	return char == '%' || unicode.In(char, unicode.Sc)
}

// Parses a phone number from the candidate using phonenumberutil.parse and
// verifies it matches the requested leniency. If parsing and verification succeed, a
// corresponding PhoneNumberMatch is returned, otherwise this method returns None.
//
// Arguments:
//
// candidate -- The candidate match.
//
// offset -- The offset of candidate within self.text.
//
// Returns the parsed and validated phone number match, or None.
func (p *PhoneNumberMatcher) parseAndVerify(candidate string, offset int) (*PhoneNumberMatch, error) {
	// Check the candidate doesn't contain any formatting which would
	// indicate that it really isn't a phone number.
	if MATCHING_BRACKETS.FindStringIndex(candidate) == nil || PUB_PAGES.FindStringIndex(candidate) != nil {
		return nil, nil
	}

	// If leniency is set to VALID or stricter, we also want to skip
	// numbers that are surrounded by Latin alphabetic characters, to
	// skip cases like abc8005001234 or 8005001234def.
	if p.leniency >= VALID {
		if offset > 0 && LEAD_PATTERN.FindStringIndex(candidate) == nil {
			// If the candidate is not at the start of the text, and does
			// not start with phone-number punctuation, check the previous
			// character
			previousChar := p.text[offset-1]
			// We return nil if it is a latin letter or an invalid
			// punctuation symbol
			if p.isInvalidPunctuationSymbol(rune(previousChar)) || unicode.IsLetter(rune(previousChar)) {
				return nil, nil
			}
		}
		lastCharIndex := offset + len(candidate)
		if lastCharIndex < len(p.text) {
			nextChar := p.text[lastCharIndex]
			if p.isInvalidPunctuationSymbol(rune(nextChar)) || unicode.IsLetter(rune(nextChar)) {
				return nil, nil
			}
		}
	}

	number, err := ParseAndKeepRawInput(candidate, p.preferredRegion)
	if err != nil {
		return nil, err
	}

	if p.leniency.Verify(number, candidate) {
		//  We used parse(keep_raw_input=True) to create this number,
		//  but for now we don't return the extra values parsed.
		//  TODO: stop clearing all values here and switch all users
		//  over to using raw_input rather than the raw_string of
		//  PhoneNumberMatch.
		match := NewPhoneNumberMatch(offset, candidate, *number)

		return &match, nil
	}

	return nil, nil
}

// Attempts to extract a match from a candidate string.
//
//  Arguments:
//
//  candidate -- The candidate text that might contain a phone number.
//
//  offset -- The offset of candidate within self.text
//
//  Returns the match found, None if none can be found
func (p *PhoneNumberMatcher) extractMatch(candidate string, offset int) *PhoneNumberMatch {
	// Skip a match that is more likely a publication page reference or a
	// date.
	if SLASH_SEPARATED_DATES.FindStringIndex(candidate) != nil {
		return nil
	}

	// Skip potential time-stamps.
	if TIME_STAMPS.FindStringIndex(candidate) != nil {
		followingText := p.text[offset+len(candidate):]
		if TIME_STAMPS_SUFFIX.FindStringIndex(followingText) != nil {
			return nil
		}
	}

	// Try to come up with a valid match given the entire candidate.
	match, _ := p.parseAndVerify(candidate, offset)
	if match != nil {
		return match
	}

	// If that failed, try to find an "inner match" -- there might be a
	// phone number within this candidate.
	return p.extractInnerMatch(candidate, offset)
}

// Attempts to extract a match from candidate if the whole candidate
// does not qualify as a match.
//
//  Arguments:
//
//  candidate -- The candidate text that might contain a phone number
//
//  offset -- The current offset of candidate within text
//
//  Returns the match found, None if none can be found
func (p *PhoneNumberMatcher) extractInnerMatch(candidate string, offset int) *PhoneNumberMatch {
	for _, possibleInnerMatch := range INNER_MATCHES {
		groupMatch := possibleInnerMatch.FindStringIndex(candidate)
		isFirstMatch := true
		index := 0
		for {
			if p.maxTries <= 0 || groupMatch == nil {
				break
			}
			if isFirstMatch {
				// We should handle any group before this one too.
				group := p.trimAfterFirstMatch(UNWANTED_END_CHAR_PATTERN, candidate[:groupMatch[0]])
				match, _ := p.parseAndVerify(group, offset)
				if match != nil {
					return match
				}
				p.maxTries--
				isFirstMatch = false
			}
			start := index + groupMatch[0]
			end := index + groupMatch[1]
			innerCandidate := candidate[start:end]
			group := p.trimAfterFirstMatch(UNWANTED_END_CHAR_PATTERN, innerCandidate)
			match, _ := p.parseAndVerify(group, offset+start)
			if match != nil {
				return match
			}

			index = start + len(innerCandidate)
			groupMatch = possibleInnerMatch.FindStringIndex(candidate[index:])
			p.maxTries--
		}
	}
	return nil
}

// Attempts to find the next subsequence in the searched sequence on or after index
// that represents a phone number. Returns the next match, None if none was found.
//
// Arguments:
//
// index -- The search index to start searching at.
//
// Returns the phone number match found, None if none can be found.
func (p *PhoneNumberMatcher) find() *PhoneNumberMatch {
	matcher := PATTERN.FindStringIndex(p.text[p.searchIndex:])
	index := 0 + p.searchIndex
	for {
		if p.maxTries <= 0 || matcher == nil {
			break
		}
		start := index + matcher[0]
		end := index + matcher[1]
		candidate := p.text[start:end]

		// Check for extra numbers at the end.
		// TODO: This is the place to start when trying to support extraction of multiple phone number
		// from split notations (+41 79 123 45 67 / 68).
		candidate = p.trimAfterFirstMatch(SECOND_NUMBER_START_PATTERN, candidate)

		match := p.extractMatch(candidate, start)
		if match != nil {
			return match
		}

		index = start + len(candidate)
		matcher = PATTERN.FindStringIndex(p.text[index:])
		p.maxTries--
	}
	return nil
}

// Indicates whether there is another match available
func (p *PhoneNumberMatcher) hasNext() bool {
	if p.state == notReady {
		p.lastMatch = p.find()
		if p.lastMatch == nil {
			p.state = done
		} else {
			p.searchIndex = p.lastMatch.end
			p.state = ready
		}
	}
	return p.state == ready
}

// Return the next match; raises Exception if no next match available
func (p *PhoneNumberMatcher) Next() (*PhoneNumberMatch, error) {
	if !p.hasNext() {
		return nil, io.EOF
	}
	// Remove from memory after use
	result := p.lastMatch
	p.lastMatch = nil
	p.state = notReady
	return result, nil
}

/*
The immutable match of a phone number within a piece of text.

Matches may be found using the find() method of PhoneNumberMatcher.

A match consists of the phone number (in .number) as well as the .start and .end offsets of the corresponding subsequence of the searched text. Use .raw_string to obtain a copy of the matched subsequence.
*/
type PhoneNumberMatch struct {
	start, end int
	rawString  string
	Number     PhoneNumber
}

func NewPhoneNumberMatch(start int, rawString string, number PhoneNumber) PhoneNumberMatch {
	return PhoneNumberMatch{
		start:     start,
		end:       start + len(rawString),
		rawString: rawString,
		Number:    number,
	}
}
