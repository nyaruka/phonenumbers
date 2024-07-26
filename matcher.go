package phonenumbers

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type PhoneNumberMatcher struct {
	PhoneNumber *PhoneNumber
	Next        *PhoneNumberMatcher
	Start       int
	End         int
}

func NewPhoneNumberMatcher(seq string) *PhoneNumberMatcher {
	// TODO(ttacon): to be implemented
	return nil
}

// NewPhoneNumberMatcherForRegion constructs a PhoneNumberMatcher for the specified text sequence and region.
// It returns a linked list of valid phone numbers found in the sequence that are valid for the given region.
// If no valid numbers are found, the function returns nil.
func NewPhoneNumberMatcherForRegion(seq string, region string) *PhoneNumberMatcher {
	var head, current *PhoneNumberMatcher
	seenNumbers := make(map[uint64]bool) // Tracks numbers to avoid duplicate entries in the list.

	// First, find all starting indices where a phone number could potentially start.
	startIndicesMatches := VALID_START_CHAR_PATTERN.FindAllIndex([]byte(seq), -1)
	startIndices := make([]int, len(startIndicesMatches))
	for i, match := range startIndicesMatches {
		startIndices[i] = match[0]
	}

	// Similarly, find all indices where a phone number should not end.
	unwantedEndIndicesMatches := UNWANTED_END_CHAR_PATTERN.FindAllIndex([]byte(seq), -1)
	unwantedEndIndices := make([]int, len(unwantedEndIndicesMatches))
	for i, match := range unwantedEndIndicesMatches {
		unwantedEndIndices[i] = match[0]
	}

	// Append the length of the sequence as an end index if not already included.
	if len(unwantedEndIndices) == 0 || unwantedEndIndices[len(unwantedEndIndices)-1] != len(seq) {
		unwantedEndIndices = append(unwantedEndIndices, len(seq)-1)
	}

	// Iterate over each possible start index.
	for i := 0; i < len(startIndices); i++ {
		for j := 0; j < len(unwantedEndIndices); j++ {
			if unwantedEndIndices[j] < startIndices[i] {
				continue // Ensure the end index is after the start index.
			}
			// Explore the sequence between the start and end indices to find valid phone numbers.
			for k := startIndices[i]; k <= unwantedEndIndices[j]; k++ {
				if k-startIndices[i] > MAX_LENGTH_COUNTRY_CODE+MAX_LENGTH_FOR_NSN {
					break // Exceeds max phone number length, stop searching this slice.
				}
				if k-startIndices[i] < MIN_LENGTH_FOR_NSN {
					continue // Does not meet min phone number length, continue searching.
				}

				phoneNumber, err := Parse(seq[startIndices[i]:k+1], region)
				if err != nil || !IsValidNumberForRegion(phoneNumber, region) {
					continue // Skip invalid numbers or parse errors.
				}

				nationalNumber := phoneNumber.GetNationalNumber()
				if _, exists := seenNumbers[nationalNumber]; exists {
					continue // Skip this number if already seen.
				}

				// Mark this phone number as seen.
				seenNumbers[nationalNumber] = true

				// Create a new matcher node and append it to the linked list.
				newMatcher := &PhoneNumberMatcher{
					PhoneNumber: phoneNumber,
					Start:       startIndices[i],
					End:         k,
				}
				if head == nil {
					head = newMatcher
					current = head
				} else {
					current.Next = newMatcher
					current = current.Next
				}
			}
			break
		}
	}
	return head
}

func ContainsOnlyValidXChars(number *PhoneNumber, candidate string) bool {
	// The characters 'x' and 'X' can be (1) a carrier code, in which
	// case they always precede the national significant number or (2)
	// an extension sign, in which case they always precede the extension
	// number. We assume a carrier code is more than 1 digit, so the first
	// case has to have more than 1 consecutive 'x' or 'X', whereas the
	// second case can only have exactly 1 'x' or 'X'. We ignore the
	// character if it appears as the last character of the string.
	for index := 0; index < len(candidate)-1; index++ {
		var charAtIndex = candidate[index]
		if charAtIndex == 'x' || charAtIndex == 'X' {
			var charAtNextIndex = candidate[index+1]
			if charAtNextIndex == 'x' || charAtNextIndex == 'X' {
				// This is the carrier code case, in which the 'X's
				// always precede the national significant number.
				index++
				if IsNumberMatchWithOneNumber(number, candidate[index:]) != NSN_MATCH {
					return false
				}
				// This is the extension sign case, in which the 'x'
				// or 'X' should always precede the extension number.
			} else if NormalizeDigitsOnly(candidate[index:]) != number.GetExtension() {
				return false
			}
		}
	}
	return true
}

func IsNationalPrefixPresentIfRequired(number *PhoneNumber) bool {
	// First, check how we deduced the country code. If it was written
	// in international format, then the national prefix is not required.
	if number.GetCountryCodeSource() != PhoneNumber_FROM_DEFAULT_COUNTRY {
		return true
	}
	var phoneNumberRegion = GetRegionCodeForCountryCode(int(number.GetCountryCode()))
	var metadata = getMetadataForRegion(phoneNumberRegion)
	if metadata == nil {
		return true
	}
	// Check if a national prefix should be present when formatting this number.
	var nationalNumber = GetNationalSignificantNumber(number)
	var formatRule = chooseFormattingPatternForNumber(
		metadata.GetNumberFormat(), nationalNumber)
	// To do this, we check that a national prefix formatting rule was
	// present and that it wasn't just the first-group symbol ($1) with
	// punctuation.
	if (formatRule != nil) && len(formatRule.GetNationalPrefixFormattingRule()) > 0 {
		if formatRule.GetNationalPrefixOptionalWhenFormatting() {
			// The national-prefix is optional in these cases, so we
			// don't need to check if it was present.
			return true
		}
		if formattingRuleHasFirstGroupOnly(
			formatRule.GetNationalPrefixFormattingRule()) {
			// National Prefix not needed for this number.
			return true
		}
		// Normalize the remainder.
		var rawInputCopy = NormalizeDigitsOnly(number.GetRawInput())
		var rawInput = NewBuilderString(rawInputCopy)
		// Check if we found a national prefix and/or carrier code at
		// the start of the raw input, and return the result.
		return maybeStripNationalPrefixAndCarrierCode(
			rawInput, metadata, NewBuilder(nil))
	}
	return true
}

func ContainsMoreThanOneSlashInNationalNumber(
	number *PhoneNumber,
	candidate string) bool {
	var firstSlash = strings.Index(candidate, "/")
	if firstSlash < 0 {
		// No slashes, this is okay.
		return false
	}
	// Now look for a second one.
	var secondSlash = strings.Index(candidate[firstSlash+1:], "/")
	if secondSlash < 0 {
		// Only one slash, this is okay.
		return false
	}

	// If the first slash is after the country calling code, this is permitted.
	var candidateHasCountryCode = (number.GetCountryCodeSource() == PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN ||
		number.GetCountryCodeSource() == PhoneNumber_FROM_NUMBER_WITHOUT_PLUS_SIGN)
	cc := strconv.Itoa(int(number.GetCountryCode()))
	if candidateHasCountryCode &&
		NormalizeDigitsOnly(candidate[0:firstSlash]) == cc {
		// Any more slashes and this is illegal.
		return strings.Contains(candidate[secondSlash+1:], "/")
	}
	return true
}

func CheckNumberGroupingIsValid(
	number *PhoneNumber,
	candidate string,
	fn func(*PhoneNumber, string, []string) bool) bool {
	// TODO(ttacon): to be implemented
	return false
}

func AllNumberGroupsRemainGrouped(
	number *PhoneNumber,
	normalizedCandidate string,
	formattedNumberGroups []string) bool {

	var fromIndex = 0
	if number.GetCountryCodeSource() != PhoneNumber_FROM_DEFAULT_COUNTRY {
		// First skip the country code if the normalized candidate contained it.
		var cc = strconv.Itoa(int(number.GetCountryCode()))
		fromIndex = strings.Index(normalizedCandidate, cc) + len(cc)
	}
	// Check each group of consecutive digits are not broken into
	// separate groupings in the normalizedCandidate string.
	for i := 0; i < len(formattedNumberGroups); i++ {
		// Fails if the substring of normalizedCandidate starting
		// from fromIndex doesn't contain the consecutive digits
		// in formattedNumberGroups[i].
		fromIndex = strings.Index(
			normalizedCandidate[fromIndex+1:], formattedNumberGroups[i])
		if fromIndex < 0 {
			return false
		}
		// Moves fromIndex forward.
		fromIndex += len(formattedNumberGroups[i])
		if i == 0 && fromIndex < len(normalizedCandidate) {
			// We are at the position right after the NDC. We get
			// the region used for formatting information based on
			// the country code in the phone number, rather than the
			// number itself, as we do not need to distinguish between
			// different countries with the same country calling code
			// and this is faster.
			var region = GetRegionCodeForCountryCode(int(number.GetCountryCode()))
			if GetNddPrefixForRegion(region, true) != "" &&
				unicode.IsDigit(rune(normalizedCandidate[fromIndex])) {
				// This means there is no formatting symbol after the
				// NDC. In this case, we only accept the number if there
				// is no formatting symbol at all in the number, except
				// for extensions. This is only important for countries
				// with national prefixes.
				var nationalSignificantNumber = GetNationalSignificantNumber(number)
				return strings.HasPrefix(
					normalizedCandidate[fromIndex-len(formattedNumberGroups[i]):],
					nationalSignificantNumber)
			}
		}
	}

	// The check here makes sure that we haven't mistakenly already
	// used the extension to match the last group of the subscriber
	// number. Note the extension cannot have formatting in-between digits.
	return strings.Contains(normalizedCandidate[fromIndex:], number.GetExtension())
}

func AllNumberGroupsAreExactlyPresent(
	number *PhoneNumber,
	normalizedCandidate string,
	formattedNumberGroups []string) bool {

	var candidateGroups = NON_DIGITS_PATTERN.FindAllString(normalizedCandidate, -1)
	// Set this to the last group, skipping it if the number has an extension.
	var candidateNumberGroupIndex = len(candidateGroups) - 2
	if number.GetExtension() != "" {
		candidateNumberGroupIndex = len(candidateGroups) - 1
	}

	// First we check if the national significant number is formatted
	// as a block. We use contains and not equals, since the national
	// significant number may be present with a prefix such as a national
	// number prefix, or the country code itself.
	if len(candidateGroups) == 1 || strings.Contains(
		candidateGroups[candidateNumberGroupIndex],
		GetNationalSignificantNumber(number)) {
		return true
	}
	// Starting from the end, go through in reverse, excluding the first
	// group, and check the candidate and number groups are the same.
	for formattedNumberGroupIndex := len(formattedNumberGroups) - 1; formattedNumberGroupIndex > 0 && candidateNumberGroupIndex >= 0; formattedNumberGroupIndex-- {
		if candidateGroups[candidateNumberGroupIndex] !=
			formattedNumberGroups[formattedNumberGroupIndex] {
			return false
		}
		candidateNumberGroupIndex--
	}
	// Now check the first group. There may be a national prefix at
	// the start, so we only check that the candidate group ends with
	// the formatted number group.
	return (candidateNumberGroupIndex >= 0 &&
		strings.HasSuffix(candidateGroups[candidateNumberGroupIndex],
			formattedNumberGroups[0]))
}

// Returns whether the given national number (a string containing only decimal digits) matches
// the national number pattern defined in the given PhoneNumberDesc message.
func MatchNationalNumber(number string, numberDesc PhoneNumberDesc, allowPrefixMatch bool) bool {
	nationalNumberPattern := numberDesc.GetNationalNumberPattern()
	// We don't want to consider it a prefix match when matching non-empty input against an empty pattern.
	if len(nationalNumberPattern) == 0 {
		return false
	}
	regex := regexFor(nationalNumberPattern)
	return match(number, regex, allowPrefixMatch)
}

func match(number string, pattern *regexp.Regexp, allowPrefixMatch bool) bool {
	ind := pattern.FindStringIndex(number)
	if len(ind) == 0 || ind[0] != 0 {
		return false
	}
	patP := `^(?:` + pattern.String() + `)$` // Strictly match
	pat := regexFor(patP)
	return pat.MatchString(number) || allowPrefixMatch
}
