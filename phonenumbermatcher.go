package phonenumbers

import (
	"io"
	"regexp"
	"strconv"
	"unicode"
)

/*
A stateful class that finds and extracts telephone numbers fom text.

Vanity numbers (phone numbers using alphabetic digits such as '1-800-SIX-FLAGS' are not found.
*/
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
	OPENING_PARENS     = "\\(\\[\uFF08\uFF3B"
	CLOSING_PARENS     = "\\)\\]\uFF09\uFF3D"
	NON_PARENS         = "[^" + OPENING_PARENS + CLOSING_PARENS + "]"
	BRACKET_PAIR_LIMIT = "{0,3}"

	LEAD_CLASS   = OPENING_PARENS + PLUS_CHARS
	LEAD_PATTERN = regexp.MustCompile(LEAD_CLASS)
	LEAD_LIMIT   = "{0,2}"

	DIGIT_BLOCK_LIMIT = 17 + 3
	DIGIT_SEQUENCE    = "\\d{1," + strconv.Itoa(DIGIT_BLOCK_LIMIT) + "}"

	PUNCTIATION_LIMIT = "{0,4}"
	PUNCTUATION       = "[" + VALID_PUNCTUATION + "]" + PUNCTIATION_LIMIT

	BLOCK_LIMIT = "{0," + strconv.Itoa(DIGIT_BLOCK_LIMIT) + "}"

	PATTERN = regexp.MustCompile("(?:" + LEAD_CLASS + PUNCTUATION + ")" + LEAD_LIMIT + DIGIT_SEQUENCE + "(?:" + PUNCTUATION + DIGIT_SEQUENCE + ")" + BLOCK_LIMIT + "(?:" + EXTN_PATTERNS_FOR_MATCHING + ")?")

	SLASH_SEPARATED_DATES = regexp.MustCompile("(?:(?:[0-3]?\\d/[01]?\\d)|(?:[01]?\\d/[0-3]?\\d))/(?:[12]\\d)?\\d{2}")
	TIME_STAMPS           = regexp.MustCompile("[12]\\d{3}[-/]?[01]\\d[-/]?[0-3]\\d +[0-2]\\d$")

	MATCHING_BRACKETS = regexp.MustCompile("(?:[" + OPENING_PARENS + "])?" + "(?:" + NON_PARENS + "+" + "[" + CLOSING_PARENS + "])?" + NON_PARENS + "+" + "(?:[" + OPENING_PARENS + "]" + NON_PARENS + "+[" + CLOSING_PARENS + "])" + BRACKET_PAIR_LIMIT + NON_PARENS + "*")

	PUB_PAGES = regexp.MustCompile("\\d{1,5}-+\\d{1,5}\\s{0,4}\\(\\d{1,4}")

	/**
	 * Patterns used to extract phone numbers from a larger phone-number-like pattern. These are
	 * ordered according to specificity. For example, white-space is last since that is frequently
	 * used in numbers, not just to separate two numbers. We have separate patterns since we don't
	 * want to break up the phone-number-like text on more than one different kind of symbol at one
	 * time, although symbols of the same type (e.g. space) can be safely grouped together.
	 *
	 * Note that if there is a match, we will always check any text found up to the first match as
	 * well.
	 */
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
)

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

func (p *PhoneNumberMatcher) parseAndVerify(candidate string, offset int) (*PhoneNumberMatch, error) {
	if MATCHING_BRACKETS.FindStringIndex(candidate) == nil || PUB_PAGES.FindStringIndex(candidate) != nil {
		return nil, nil
	}

	if p.leniency >= VALID {
		if offset > 0 && LEAD_PATTERN.FindStringIndex(candidate) == nil {
			previousChar := []rune(p.text)[offset-1]
			if p.isInvalidPunctuationSymbol(previousChar) || unicode.IsLetter(previousChar) {
				return nil, nil
			}
		}
		lastCharIndex := offset + len(candidate)
		if lastCharIndex < len(p.text) {
			nextChar := []rune(p.text)[lastCharIndex]
			if p.isInvalidPunctuationSymbol(nextChar) || unicode.IsLetter(nextChar) {
				return nil, nil
			}
		}
	}

	number, err := ParseAndKeepRawInput(candidate, p.preferredRegion)
	if err != nil {
		return nil, err
	}

	if p.leniency.Verify(number, candidate) {
		match := NewPhoneNumberMatch(offset, candidate, *number)

		return &match, nil
	}

	return nil, nil
}

func (p *PhoneNumberMatcher) extractMatch(candidate string, offset int) *PhoneNumberMatch {
	if SLASH_SEPARATED_DATES.FindStringIndex(candidate) != nil {
		return nil
	}

	if TIME_STAMPS.FindStringIndex(candidate) != nil {
		return nil
	}

	match, _ := p.parseAndVerify(candidate, offset)
	if match != nil {
		return match
	}

	return p.extractInnerMatch(candidate, offset)
}

func (p *PhoneNumberMatcher) extractInnerMatch(candidate string, offset int) *PhoneNumberMatch {
	for _, possibleInnerMatch := range INNER_MATCHES {
		groupMatch := possibleInnerMatch.FindStringIndex(candidate)
		isFirstMatch := true
		for {
			if groupMatch == nil || p.maxTries == 0 {
				break
			}
			if isFirstMatch {
				group := p.trimAfterFirstMatch(UNWANTED_END_CHAR_PATTERN, candidate[:groupMatch[0]])
				match, _ := p.parseAndVerify(group, offset+groupMatch[0])
				if match != nil {
					return match
				}
				p.maxTries--
				isFirstMatch = false
			}
			group := p.trimAfterFirstMatch(UNWANTED_END_CHAR_PATTERN, candidate[groupMatch[0]:groupMatch[1]])
			match, _ := p.parseAndVerify(group, offset+groupMatch[0])
			if match != nil {
				return match
			}
			p.maxTries--
			groupMatch = possibleInnerMatch.FindStringIndex(candidate[groupMatch[1]:])
		}
	}
	return nil
}

func (p *PhoneNumberMatcher) find() *PhoneNumberMatch {
	matcher := PATTERN.FindStringIndex(p.text[p.searchIndex:])
	index := 0
	for {
		if p.maxTries > 0 && matcher == nil {
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
	number     PhoneNumber
}

func NewPhoneNumberMatch(start int, rawString string, number PhoneNumber) PhoneNumberMatch {
	return PhoneNumberMatch{
		start:     start,
		end:       start + len(rawString),
		rawString: rawString,
		number:    number,
	}
}
