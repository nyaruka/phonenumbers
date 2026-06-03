// Port of AsYouTypeFormatter.java from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package phonenumbers

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nyaruka/phonenumbers/v2/internal/regexcache"
	"github.com/nyaruka/phonenumbers/v2/internal/stringbuilder"
	"google.golang.org/protobuf/proto"
)

// AsYouTypeFormatter formats phone numbers as they are entered. It is a port of
// libphonenumber's AsYouTypeFormatter (Java reference implementation).
//
// An AsYouTypeFormatter is obtained from GetAsYouTypeFormatter. After that,
// digits can be added by calling InputDigit on the instance, and the partially
// formatted phone number is returned each time a digit is added. Clear can be
// called before formatting a new number.
//
// See AsYouTypeFormatter's tests for more details on how the formatter is to be
// used.
type AsYouTypeFormatter struct {
	currentOutput string
	// formattingTemplate holds the current template with not-yet-entered digits
	// represented by the DIGIT_PLACEHOLDER rune. It is kept as a []rune (rather
	// than a StringBuilder) because the placeholder is a multi-byte rune and the
	// template is manipulated by character index, where rune indices match Java's
	// UTF-16 char indices for every (BMP) character a phone number can contain.
	formattingTemplate []rune
	// currentFormattingPattern is the pattern from the NumberFormat currently used
	// to create formattingTemplate.
	currentFormattingPattern      string
	accruedInput                  *stringbuilder.Builder
	accruedInputWithoutFormatting *stringbuilder.Builder
	// ableToFormat indicates whether the formatter is currently doing formatting.
	ableToFormat bool
	// inputHasFormatting is set to true when users enter their own formatting. The
	// formatter will do no formatting at all when this is true.
	inputHasFormatting bool
	// isCompleteNumber is set to true when we know the user is entering a full
	// national significant number, since we have either detected a national prefix
	// or an international dialing prefix. When this is true, we will no longer use
	// local number formatting patterns.
	isCompleteNumber              bool
	isExpectingCountryCallingCode bool
	defaultCountry                string

	defaultMetadata *PhoneMetadata
	currentMetadata *PhoneMetadata

	lastMatchPosition int
	// originalPosition is the position of a digit upon which
	// InputDigitAndRememberPosition was most recently invoked, as found in the
	// original sequence of characters the user entered.
	originalPosition int
	// positionToRemember is the position of a digit upon which
	// InputDigitAndRememberPosition was most recently invoked, as found in
	// accruedInputWithoutFormatting.
	positionToRemember int
	// prefixBeforeNationalNumber contains anything that has been entered so far
	// preceding the national significant number, and it is formatted (e.g. with
	// space inserted). For example, this can contain IDD, country code, and/or
	// NDD, etc.
	prefixBeforeNationalNumber        *stringbuilder.Builder
	shouldAddSpaceAfterNationalPrefix bool
	// extractedNationalPrefix contains the national prefix that has been extracted.
	// It contains only digits without formatting.
	extractedNationalPrefix string
	nationalNumber          *stringbuilder.Builder
	possibleFormats         []*NumberFormat
}

// SEPARATOR_BEFORE_NATIONAL_NUMBER is the character used when appropriate to
// separate a prefix, such as a long NDD or a country calling code, from the
// national number.
const SEPARATOR_BEFORE_NATIONAL_NUMBER = ' '

// MIN_LEADING_DIGITS_LENGTH is the minimum length of national number accrued
// that is required to trigger the formatter. The first element of the
// leadingDigitsPattern of each numberFormat contains a regular expression that
// matches up to this number of digits.
const MIN_LEADING_DIGITS_LENGTH = 3

// DIGIT_PLACEHOLDER represents the digits that have not been entered yet, using
//  , the punctuation space.
const DIGIT_PLACEHOLDER = " "

const digitPlaceholderRune = ' '

var (
	// EMPTY_METADATA is used as a default instance of metadata. It allows the
	// formatter to function with an incorrect region code, even if formatting only
	// works for numbers specified with "+".
	EMPTY_METADATA = &PhoneMetadata{
		Id:                  proto.String("<ignored>"),
		InternationalPrefix: proto.String("NA"),
	}

	// ELIGIBLE_FORMAT_PATTERN determines if a numberFormat under availableFormats is
	// eligible to be used by the formatter. It is eligible when the format element
	// under numberFormat contains groups of the dollar sign followed by a single
	// digit, separated by valid phone number punctuation. This prevents invalid
	// punctuation (such as the star sign in Israeli star numbers) getting into the
	// output of the formatter. We require that the first group is present in the
	// output pattern to ensure no data is lost while formatting; when we format as
	// you type, this should always be the case. Anchored to require a full match,
	// mirroring Java's Matcher.matches().
	ELIGIBLE_FORMAT_PATTERN = regexp.MustCompile("^(?:[" + VALID_PUNCTUATION + "]*" +
		"\\$1" + "[" + VALID_PUNCTUATION + "]*(?:\\$\\d" + "[" + VALID_PUNCTUATION + "]*)*)$")

	// NATIONAL_PREFIX_SEPARATORS_PATTERN is a set of characters that, if found in a
	// national prefix formatting rule, are an indicator to us that we should
	// separate the national prefix from the number when formatting.
	NATIONAL_PREFIX_SEPARATORS_PATTERN = regexp.MustCompile("[- ]")
)

// GetAsYouTypeFormatter returns an AsYouTypeFormatter for the specific region.
func GetAsYouTypeFormatter(regionCode string) *AsYouTypeFormatter {
	return newAsYouTypeFormatter(regionCode)
}

// newAsYouTypeFormatter constructs an as-you-type formatter for the given region.
func newAsYouTypeFormatter(regionCode string) *AsYouTypeFormatter {
	aytf := &AsYouTypeFormatter{
		ableToFormat:                  true,
		defaultCountry:                regionCode,
		accruedInput:                  stringbuilder.New(nil),
		accruedInputWithoutFormatting: stringbuilder.New(nil),
		prefixBeforeNationalNumber:    stringbuilder.New(nil),
		nationalNumber:                stringbuilder.New(nil),
	}
	aytf.currentMetadata = aytf.getMetadataForRegion(regionCode)
	aytf.defaultMetadata = aytf.currentMetadata
	return aytf
}

// getMetadataForRegion returns the metadata needed by this formatter. It is the
// same for all regions sharing the same country calling code, so we return the
// metadata for the "main" region for this country calling code.
func (aytf *AsYouTypeFormatter) getMetadataForRegion(regionCode string) *PhoneMetadata {
	countryCallingCode := GetCountryCodeForRegion(regionCode)
	mainCountry := GetRegionCodeForCountryCode(countryCallingCode)
	metadata := getMetadataForRegion(mainCountry)
	if metadata != nil {
		return metadata
	}
	return EMPTY_METADATA
}

// maybeCreateNewTemplate returns true if a new template is created as opposed to
// reusing the existing template.
func (aytf *AsYouTypeFormatter) maybeCreateNewTemplate() bool {
	// When there are multiple available formats, the formatter uses the first format
	// where a formatting template could be created.
	i := 0
	for i < len(aytf.possibleFormats) {
		numberFormat := aytf.possibleFormats[i]
		pattern := numberFormat.GetPattern()
		if aytf.currentFormattingPattern == pattern {
			return false
		}
		if aytf.createFormattingTemplate(numberFormat) {
			aytf.currentFormattingPattern = pattern
			aytf.shouldAddSpaceAfterNationalPrefix =
				NATIONAL_PREFIX_SEPARATORS_PATTERN.MatchString(numberFormat.GetNationalPrefixFormattingRule())
			// With a new formatting template, the matched position using the old template
			// needs to be reset.
			aytf.lastMatchPosition = 0
			return true
		}
		// Remove the current number format from possibleFormats.
		aytf.possibleFormats = append(aytf.possibleFormats[:i], aytf.possibleFormats[i+1:]...)
	}
	aytf.ableToFormat = false
	return false
}

func (aytf *AsYouTypeFormatter) getAvailableFormats(leadingDigits string) {
	// First decide whether we should use international or national number rules.
	isInternationalNumber := aytf.isCompleteNumber && len(aytf.extractedNationalPrefix) == 0
	var formatList []*NumberFormat
	if isInternationalNumber && len(aytf.currentMetadata.GetIntlNumberFormat()) > 0 {
		formatList = aytf.currentMetadata.GetIntlNumberFormat()
	} else {
		formatList = aytf.currentMetadata.GetNumberFormat()
	}
	for _, format := range formatList {
		// Discard a few formats that we know are not relevant based on the presence of
		// the national prefix.
		if len(aytf.extractedNationalPrefix) > 0 &&
			formattingRuleHasFirstGroupOnly(format.GetNationalPrefixFormattingRule()) &&
			!format.GetNationalPrefixOptionalWhenFormatting() &&
			len(format.GetDomesticCarrierCodeFormattingRule()) == 0 {
			// If it is a national number that had a national prefix, any rules that aren't
			// valid with a national prefix should be excluded. A rule that has a carrier-code
			// formatting rule is kept since the national prefix might actually be an extracted
			// carrier code - we don't distinguish between these when extracting it in the AYTF.
			continue
		} else if len(aytf.extractedNationalPrefix) == 0 &&
			!aytf.isCompleteNumber &&
			!formattingRuleHasFirstGroupOnly(format.GetNationalPrefixFormattingRule()) &&
			!format.GetNationalPrefixOptionalWhenFormatting() {
			// This number was entered without a national prefix, and this formatting rule
			// requires one, so we discard it.
			continue
		}
		if ELIGIBLE_FORMAT_PATTERN.MatchString(format.GetFormat()) {
			aytf.possibleFormats = append(aytf.possibleFormats, format)
		}
	}
	aytf.narrowDownPossibleFormats(leadingDigits)
}

func (aytf *AsYouTypeFormatter) narrowDownPossibleFormats(leadingDigits string) {
	indexOfLeadingDigitsPattern := len(leadingDigits) - MIN_LEADING_DIGITS_LENGTH
	i := 0
	for i < len(aytf.possibleFormats) {
		format := aytf.possibleFormats[i]
		if len(format.GetLeadingDigitsPattern()) == 0 {
			// Keep everything that isn't restricted by leading digits.
			i++
			continue
		}
		lastLeadingDigitsPattern := indexOfLeadingDigitsPattern
		if count := len(format.GetLeadingDigitsPattern()) - 1; count < lastLeadingDigitsPattern {
			lastLeadingDigitsPattern = count
		}
		leadingDigitsPattern := regexcache.For("^(?:" + format.GetLeadingDigitsPattern()[lastLeadingDigitsPattern] + ")")
		if !leadingDigitsPattern.MatchString(leadingDigits) {
			aytf.possibleFormats = append(aytf.possibleFormats[:i], aytf.possibleFormats[i+1:]...)
		} else {
			i++
		}
	}
}

func (aytf *AsYouTypeFormatter) createFormattingTemplate(format *NumberFormat) bool {
	numberPattern := format.GetPattern()
	aytf.formattingTemplate = aytf.formattingTemplate[:0]
	tempTemplate := aytf.getFormattingTemplate(numberPattern, format.GetFormat())
	if len(tempTemplate) > 0 {
		aytf.formattingTemplate = append(aytf.formattingTemplate, []rune(tempTemplate)...)
		return true
	}
	return false
}

// getFormattingTemplate gets a formatting template which can be used to
// efficiently format a partial number where digits are added one by one.
func (aytf *AsYouTypeFormatter) getFormattingTemplate(numberPattern, numberFormat string) string {
	// Creates a phone number consisting only of the digit 9 that matches the
	// numberPattern by applying the pattern to the longestPhoneNumber string.
	longestPhoneNumber := "999999999999999"
	m := regexcache.For(numberPattern)
	aPhoneNumber := m.FindString(longestPhoneNumber) // this will always succeed
	// No formatting template can be created if the number of digits entered so far is
	// longer than the maximum the current formatting rule can accommodate.
	if len(aPhoneNumber) < aytf.nationalNumber.Len() {
		return ""
	}
	// Formats the number according to numberFormat.
	template := m.ReplaceAllString(aPhoneNumber, numberFormat)
	// Replaces each digit with character DIGIT_PLACEHOLDER.
	template = strings.ReplaceAll(template, "9", DIGIT_PLACEHOLDER)
	return template
}

// Clear clears the internal state of the formatter, so it can be reused.
func (aytf *AsYouTypeFormatter) Clear() {
	aytf.currentOutput = ""
	aytf.accruedInput.Reset()
	aytf.accruedInputWithoutFormatting.Reset()
	aytf.formattingTemplate = aytf.formattingTemplate[:0]
	aytf.lastMatchPosition = 0
	aytf.currentFormattingPattern = ""
	aytf.prefixBeforeNationalNumber.Reset()
	aytf.extractedNationalPrefix = ""
	aytf.nationalNumber.Reset()
	aytf.ableToFormat = true
	aytf.inputHasFormatting = false
	aytf.positionToRemember = 0
	aytf.originalPosition = 0
	aytf.isCompleteNumber = false
	aytf.isExpectingCountryCallingCode = false
	aytf.possibleFormats = aytf.possibleFormats[:0]
	aytf.shouldAddSpaceAfterNationalPrefix = false
	if aytf.currentMetadata != aytf.defaultMetadata {
		aytf.currentMetadata = aytf.getMetadataForRegion(aytf.defaultCountry)
	}
}

// InputDigit formats a phone number on-the-fly as each digit is entered.
//
// nextChar is the most recently entered digit of a phone number. Formatting
// characters are allowed, but as soon as they are encountered this method formats
// the number as entered and not "as you type" anymore. Full width digits and
// Arabic-indic digits are allowed, and will be shown as they are. It returns the
// partially formatted phone number.
func (aytf *AsYouTypeFormatter) InputDigit(nextChar rune) string {
	aytf.currentOutput = aytf.inputDigitWithOptionToRememberPosition(nextChar, false)
	return aytf.currentOutput
}

// InputDigitAndRememberPosition is the same as InputDigit, but remembers the
// position where nextChar is inserted, so that it can be retrieved later using
// GetRememberedPosition. The remembered position will be automatically adjusted
// if additional formatting characters are later inserted/removed in front of
// nextChar.
func (aytf *AsYouTypeFormatter) InputDigitAndRememberPosition(nextChar rune) string {
	aytf.currentOutput = aytf.inputDigitWithOptionToRememberPosition(nextChar, true)
	return aytf.currentOutput
}

func (aytf *AsYouTypeFormatter) inputDigitWithOptionToRememberPosition(nextChar rune, rememberPosition bool) string {
	aytf.accruedInput.WriteRune(nextChar)
	if rememberPosition {
		aytf.originalPosition = utf8.RuneCount(aytf.accruedInput.Bytes())
	}
	// We do formatting on-the-fly only when each character entered is either a digit,
	// or a plus sign (accepted at the start of the number only).
	if !aytf.isDigitOrLeadingPlusSign(nextChar) {
		aytf.ableToFormat = false
		aytf.inputHasFormatting = true
	} else {
		nextChar = aytf.normalizeAndAccrueDigitsAndPlusSign(nextChar, rememberPosition)
	}
	if !aytf.ableToFormat {
		// When we are unable to format because of reasons other than that formatting chars
		// have been entered, it can be due to really long IDDs or NDDs. If that is the case,
		// we might be able to do formatting again after extracting them.
		if aytf.inputHasFormatting {
			return aytf.accruedInput.String()
		} else if aytf.attemptToExtractIdd() {
			if aytf.attemptToExtractCountryCallingCode() {
				return aytf.attemptToChoosePatternWithPrefixExtracted()
			}
		} else if aytf.ableToExtractLongerNdd() {
			// Add an additional space to separate long NDD and national significant number for
			// readability. We don't set shouldAddSpaceAfterNationalPrefix to true, since we don't
			// want this to change later when we choose formatting templates.
			aytf.prefixBeforeNationalNumber.WriteByte(byte(SEPARATOR_BEFORE_NATIONAL_NUMBER))
			return aytf.attemptToChoosePatternWithPrefixExtracted()
		}
		return aytf.accruedInput.String()
	}

	// We start to attempt to format only when at least MIN_LEADING_DIGITS_LENGTH
	// digits (the plus sign is counted as a digit as well for this purpose) have been
	// entered.
	switch aytf.accruedInputWithoutFormatting.Len() {
	case 0, 1, 2:
		return aytf.accruedInput.String()
	case 3:
		if aytf.attemptToExtractIdd() {
			aytf.isExpectingCountryCallingCode = true
		} else { // No IDD or plus sign is found, might be entering in national format.
			aytf.extractedNationalPrefix = aytf.removeNationalPrefixFromNationalNumber()
			return aytf.attemptToChooseFormattingPattern()
		}
		fallthrough
	default:
		if aytf.isExpectingCountryCallingCode {
			if aytf.attemptToExtractCountryCallingCode() {
				aytf.isExpectingCountryCallingCode = false
			}
			return aytf.prefixBeforeNationalNumber.String() + aytf.nationalNumber.String()
		}
		if len(aytf.possibleFormats) > 0 { // The formatting patterns are already chosen.
			tempNationalNumber := aytf.inputDigitHelper(nextChar)
			// See if the accrued digits can be formatted properly already. If not, use the
			// results from inputDigitHelper, which does formatting based on the formatting
			// pattern chosen.
			formattedNumber := aytf.attemptToFormatAccruedDigits()
			if len(formattedNumber) > 0 {
				return formattedNumber
			}
			aytf.narrowDownPossibleFormats(aytf.nationalNumber.String())
			if aytf.maybeCreateNewTemplate() {
				return aytf.inputAccruedNationalNumber()
			}
			if aytf.ableToFormat {
				return aytf.appendNationalNumber(tempNationalNumber)
			}
			return aytf.accruedInput.String()
		}
		return aytf.attemptToChooseFormattingPattern()
	}
}

func (aytf *AsYouTypeFormatter) attemptToChoosePatternWithPrefixExtracted() string {
	aytf.ableToFormat = true
	aytf.isExpectingCountryCallingCode = false
	aytf.possibleFormats = aytf.possibleFormats[:0]
	aytf.lastMatchPosition = 0
	aytf.formattingTemplate = aytf.formattingTemplate[:0]
	aytf.currentFormattingPattern = ""
	return aytf.attemptToChooseFormattingPattern()
}

// getExtractedNationalPrefix is visible for testing.
func (aytf *AsYouTypeFormatter) getExtractedNationalPrefix() string {
	return aytf.extractedNationalPrefix
}

// ableToExtractLongerNdd handles the case where some national prefixes are a
// substring of others. If extracting the shorter NDD doesn't result in a number
// we can format, we try to see if we can extract a longer version here.
func (aytf *AsYouTypeFormatter) ableToExtractLongerNdd() bool {
	if len(aytf.extractedNationalPrefix) > 0 {
		// Put the extracted NDD back to the national number before attempting to extract a
		// new NDD.
		aytf.nationalNumber.InsertString(0, aytf.extractedNationalPrefix)
		// Remove the previously extracted NDD from prefixBeforeNationalNumber. We cannot
		// simply set it to empty string because people sometimes incorrectly enter national
		// prefix after the country code, e.g. +44 (0)20-1234-5678.
		indexOfPreviousNdd := aytf.prefixBeforeNationalNumber.LastIndexOf(aytf.extractedNationalPrefix)
		aytf.prefixBeforeNationalNumber.SetLength(indexOfPreviousNdd)
	}
	return aytf.extractedNationalPrefix != aytf.removeNationalPrefixFromNationalNumber()
}

func (aytf *AsYouTypeFormatter) isDigitOrLeadingPlusSign(nextChar rune) bool {
	return unicode.IsDigit(nextChar) ||
		(utf8.RuneCount(aytf.accruedInput.Bytes()) == 1 &&
			PLUS_CHARS_PATTERN.MatchString(string(nextChar)))
}

// attemptToFormatAccruedDigits checks to see if there is an exact pattern match
// for these digits. If so, we should use this instead of any other formatting
// template whose leadingDigitsPattern also matches the input.
func (aytf *AsYouTypeFormatter) attemptToFormatAccruedDigits() string {
	for _, numberFormat := range aytf.possibleFormats {
		m := regexcache.For("^(?:" + numberFormat.GetPattern() + ")$")
		if m.MatchString(aytf.nationalNumber.String()) {
			aytf.shouldAddSpaceAfterNationalPrefix =
				NATIONAL_PREFIX_SEPARATORS_PATTERN.MatchString(numberFormat.GetNationalPrefixFormattingRule())
			formattedNumber := m.ReplaceAllString(aytf.nationalNumber.String(), numberFormat.GetFormat())
			// Check that we did not remove nor add any extra digits when we matched this
			// formatting pattern. This usually happens after we entered the last digit during
			// AYTF. Eg: In case of MX, we swallow mobile token (1) when formatted but AYTF should
			// retain all the number entered and not change in order to match a format (of same
			// leading digits and length) display in that way.
			fullOutput := aytf.appendNationalNumber(formattedNumber)
			formattedNumberDigitsOnly := normalizeDiallableCharsOnly(fullOutput)
			if formattedNumberDigitsOnly == aytf.accruedInputWithoutFormatting.String() {
				// If it's the same (i.e entered number and format is same), then it's safe to
				// return this in formatted number as nothing is lost / added.
				return fullOutput
			}
		}
	}
	return ""
}

// GetRememberedPosition returns the current position in the partially formatted
// phone number of the character which was previously passed in as the parameter
// of InputDigitAndRememberPosition.
func (aytf *AsYouTypeFormatter) GetRememberedPosition() int {
	if !aytf.ableToFormat {
		return aytf.originalPosition
	}
	// Positions are tracked as rune counts, which match Java's UTF-16 char positions
	// for every (BMP) character a phone number can contain.
	accruedInputRunes := []rune(aytf.accruedInputWithoutFormatting.String())
	currentOutputRunes := []rune(aytf.currentOutput)
	accruedInputIndex := 0
	currentOutputIndex := 0
	for accruedInputIndex < aytf.positionToRemember && currentOutputIndex < len(currentOutputRunes) {
		if accruedInputRunes[accruedInputIndex] == currentOutputRunes[currentOutputIndex] {
			accruedInputIndex++
		}
		currentOutputIndex++
	}
	return currentOutputIndex
}

// appendNationalNumber combines the national number with any prefix (IDD/+ and
// country code or national prefix) that was collected. A space will be inserted
// between them if the current formatting template indicates this to be suitable.
func (aytf *AsYouTypeFormatter) appendNationalNumber(nationalNumber string) string {
	prefixBeforeNationalNumberLength := aytf.prefixBeforeNationalNumber.Len()
	if aytf.shouldAddSpaceAfterNationalPrefix && prefixBeforeNationalNumberLength > 0 &&
		aytf.prefixBeforeNationalNumber.CharAt(prefixBeforeNationalNumberLength-1) != byte(SEPARATOR_BEFORE_NATIONAL_NUMBER) {
		// We want to add a space after the national prefix if the national prefix formatting
		// rule indicates that this would normally be done, with the exception of the case
		// where we already appended a space because the NDD was surprisingly long.
		return aytf.prefixBeforeNationalNumber.String() + string(SEPARATOR_BEFORE_NATIONAL_NUMBER) + nationalNumber
	}
	return aytf.prefixBeforeNationalNumber.String() + nationalNumber
}

// attemptToChooseFormattingPattern attempts to set the formatting template and
// returns a string which contains the formatted version of the digits entered so
// far.
func (aytf *AsYouTypeFormatter) attemptToChooseFormattingPattern() string {
	// We start to attempt to format only when at least MIN_LEADING_DIGITS_LENGTH
	// digits of national number (excluding national prefix) have been entered.
	if aytf.nationalNumber.Len() >= MIN_LEADING_DIGITS_LENGTH {
		aytf.getAvailableFormats(aytf.nationalNumber.String())
		// See if the accrued digits can be formatted properly already.
		formattedNumber := aytf.attemptToFormatAccruedDigits()
		if len(formattedNumber) > 0 {
			return formattedNumber
		}
		if aytf.maybeCreateNewTemplate() {
			return aytf.inputAccruedNationalNumber()
		}
		return aytf.accruedInput.String()
	}
	return aytf.appendNationalNumber(aytf.nationalNumber.String())
}

// inputAccruedNationalNumber invokes inputDigitHelper on each digit of the
// national number accrued, and returns a formatted string in the end.
func (aytf *AsYouTypeFormatter) inputAccruedNationalNumber() string {
	lengthOfNationalNumber := aytf.nationalNumber.Len()
	if lengthOfNationalNumber > 0 {
		tempNationalNumber := ""
		for i := 0; i < lengthOfNationalNumber; i++ {
			tempNationalNumber = aytf.inputDigitHelper(rune(aytf.nationalNumber.CharAt(i)))
		}
		if aytf.ableToFormat {
			return aytf.appendNationalNumber(tempNationalNumber)
		}
		return aytf.accruedInput.String()
	}
	return aytf.prefixBeforeNationalNumber.String()
}

// isNanpaNumberWithNationalPrefix returns true if the current country is a NANPA
// country and the national number begins with the national prefix.
func (aytf *AsYouTypeFormatter) isNanpaNumberWithNationalPrefix() bool {
	// For NANPA numbers beginning with 1[2-9], treat the 1 as the national prefix.
	// The reason is that national significant numbers in NANPA always start with [2-9]
	// after the national prefix. Numbers beginning with 1[01] can only be
	// short/emergency numbers, which don't need the national prefix.
	return aytf.currentMetadata.GetCountryCode() == 1 && aytf.nationalNumber.Len() >= 2 &&
		aytf.nationalNumber.CharAt(0) == '1' &&
		aytf.nationalNumber.CharAt(1) != '0' && aytf.nationalNumber.CharAt(1) != '1'
}

// removeNationalPrefixFromNationalNumber returns the national prefix extracted,
// or an empty string if it is not present.
func (aytf *AsYouTypeFormatter) removeNationalPrefixFromNationalNumber() string {
	startOfNationalNumber := 0
	if aytf.isNanpaNumberWithNationalPrefix() {
		startOfNationalNumber = 1
		aytf.prefixBeforeNationalNumber.WriteByte('1')
		aytf.prefixBeforeNationalNumber.WriteByte(byte(SEPARATOR_BEFORE_NATIONAL_NUMBER))
		aytf.isCompleteNumber = true
	} else if len(aytf.currentMetadata.GetNationalPrefixForParsing()) > 0 {
		nationalPrefixForParsing := regexcache.For("^(?:" + aytf.currentMetadata.GetNationalPrefixForParsing() + ")")
		// Since some national prefix patterns are entirely optional, check that a national
		// prefix could actually be extracted.
		loc := nationalPrefixForParsing.FindStringIndex(aytf.nationalNumber.String())
		if loc != nil && loc[1] > 0 {
			// When the national prefix is detected, we use international formatting rules
			// instead of national ones, because national formatting rules could contain local
			// formatting rules for numbers entered without area code.
			aytf.isCompleteNumber = true
			startOfNationalNumber = loc[1]
			aytf.prefixBeforeNationalNumber.WriteString(aytf.nationalNumber.Substring(0, startOfNationalNumber))
		}
	}
	nationalPrefix := aytf.nationalNumber.Substring(0, startOfNationalNumber)
	aytf.nationalNumber.Delete(0, startOfNationalNumber)
	return nationalPrefix
}

// attemptToExtractIdd extracts IDD and plus sign to prefixBeforeNationalNumber
// when they are available, and places the remaining input into nationalNumber. It
// returns true when accruedInputWithoutFormatting begins with the plus sign or
// valid IDD for defaultCountry.
func (aytf *AsYouTypeFormatter) attemptToExtractIdd() bool {
	internationalPrefix := regexcache.For("^(?:" + "\\" + string(PLUS_SIGN) + "|" +
		aytf.currentMetadata.GetInternationalPrefix() + ")")
	accruedInputWithoutFormatting := aytf.accruedInputWithoutFormatting.String()
	loc := internationalPrefix.FindStringIndex(accruedInputWithoutFormatting)
	if loc != nil {
		aytf.isCompleteNumber = true
		startOfCountryCallingCode := loc[1]
		aytf.nationalNumber.Reset()
		aytf.nationalNumber.WriteString(accruedInputWithoutFormatting[startOfCountryCallingCode:])
		aytf.prefixBeforeNationalNumber.Reset()
		aytf.prefixBeforeNationalNumber.WriteString(accruedInputWithoutFormatting[:startOfCountryCallingCode])
		if aytf.accruedInputWithoutFormatting.CharAt(0) != byte(PLUS_SIGN) {
			aytf.prefixBeforeNationalNumber.WriteByte(byte(SEPARATOR_BEFORE_NATIONAL_NUMBER))
		}
		return true
	}
	return false
}

// attemptToExtractCountryCallingCode extracts the country calling code from the
// beginning of nationalNumber to prefixBeforeNationalNumber when available, and
// places the remaining input into nationalNumber. It returns true when a valid
// country calling code can be found.
func (aytf *AsYouTypeFormatter) attemptToExtractCountryCallingCode() bool {
	if aytf.nationalNumber.Len() == 0 {
		return false
	}
	numberWithoutCountryCallingCode := stringbuilder.New(nil)
	countryCode := extractCountryCode(aytf.nationalNumber, numberWithoutCountryCallingCode)
	if countryCode == 0 {
		return false
	}
	aytf.nationalNumber.Reset()
	aytf.nationalNumber.Write(numberWithoutCountryCallingCode.Bytes())
	newRegionCode := GetRegionCodeForCountryCode(countryCode)
	if REGION_CODE_FOR_NON_GEO_ENTITY == newRegionCode {
		aytf.currentMetadata = getMetadataForNonGeographicalRegion(countryCode)
	} else if newRegionCode != aytf.defaultCountry {
		aytf.currentMetadata = aytf.getMetadataForRegion(newRegionCode)
	}
	countryCodeString := strconv.Itoa(countryCode)
	aytf.prefixBeforeNationalNumber.WriteString(countryCodeString)
	aytf.prefixBeforeNationalNumber.WriteByte(byte(SEPARATOR_BEFORE_NATIONAL_NUMBER))
	// When we have successfully extracted the IDD, the previously extracted NDD should
	// be cleared because it is no longer valid.
	aytf.extractedNationalPrefix = ""
	return true
}

// normalizeAndAccrueDigitsAndPlusSign accrues digits and the plus sign to
// accruedInputWithoutFormatting for later use. If nextChar contains a digit in
// non-ASCII format (e.g. the full-width version of digits), it is first
// normalized to the ASCII version. The return value is nextChar itself, or its
// normalized version, if nextChar is a digit in non-ASCII format. This method
// assumes its input is either a digit or the plus sign.
func (aytf *AsYouTypeFormatter) normalizeAndAccrueDigitsAndPlusSign(nextChar rune, rememberPosition bool) rune {
	var normalizedChar rune
	if nextChar == PLUS_SIGN {
		normalizedChar = nextChar
		aytf.accruedInputWithoutFormatting.WriteRune(nextChar)
	} else {
		if v, ok := arabicIndicNumberals[nextChar]; ok {
			normalizedChar = v
		} else {
			normalizedChar = nextChar
		}
		aytf.accruedInputWithoutFormatting.WriteRune(normalizedChar)
		aytf.nationalNumber.WriteRune(normalizedChar)
	}
	if rememberPosition {
		aytf.positionToRemember = utf8.RuneCount(aytf.accruedInputWithoutFormatting.Bytes())
	}
	return normalizedChar
}

func (aytf *AsYouTypeFormatter) inputDigitHelper(nextChar rune) string {
	// Note that formattingTemplate is not guaranteed to have a value, it could be
	// empty, e.g. when the next digit is entered after extracting an IDD or NDD.
	// Find the next DIGIT_PLACEHOLDER at or after lastMatchPosition; because every
	// placeholder before lastMatchPosition has already been filled in with a digit,
	// this is also the first remaining placeholder overall.
	idx := -1
	for i := aytf.lastMatchPosition; i < len(aytf.formattingTemplate); i++ {
		if aytf.formattingTemplate[i] == digitPlaceholderRune {
			idx = i
			break
		}
	}
	if idx >= 0 {
		aytf.formattingTemplate[idx] = nextChar
		aytf.lastMatchPosition = idx
		return string(aytf.formattingTemplate[:aytf.lastMatchPosition+1])
	}
	if len(aytf.possibleFormats) == 1 {
		// More digits are entered than we could handle, and there are no other valid
		// patterns to try.
		aytf.ableToFormat = false
	} // else, we just reset the formatting pattern.
	aytf.currentFormattingPattern = ""
	return aytf.accruedInput.String()
}
