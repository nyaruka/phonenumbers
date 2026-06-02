package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberMatcherTest.java,
// reconciled against v9.0.32. Runs against the synthetic test metadata via
// useTestMetadata(t) (so it must not run with t.Parallel()), the analogue of
// upstream's TestMetadataTestCase.
//
// Notes on the Go port:
//   - The upstream iterator (hasNext/next) is exercised through the internal
//     *phoneNumberMatcher, which is exactly what the public FindNumbers iter.Seq
//     wraps. "Collect all matches" cases and ensureTermination go through the
//     public FindNumbers / FindNumbersWithLeniency API.
//   - Offsets (Start/End) are byte offsets, so prefix lengths in the helpers use
//     len() (bytes) and substring extraction is text[start:end].
//   - testRemovalNotSupported has no Go analogue (range-over-func has no remove),
//     so it is intentionally omitted.

import (
	"iter"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// --- helpers -----------------------------------------------------------------

func ccSource(s PhoneNumber_CountryCodeSource) *PhoneNumber_CountryCodeSource { return s.Enum() }

// findMatcher mirrors upstream phoneUtil.findNumbers(text, region).iterator().
func findMatcher(text, region string) *phoneNumberMatcher {
	return newPhoneNumberMatcher(text, region, VALID, math.MaxInt)
}

// findMatcherLeniency mirrors phoneUtil.findNumbers(text, region, leniency, maxTries).iterator().
func findMatcherLeniency(text, region string, leniency Leniency, maxTries int) *phoneNumberMatcher {
	return newPhoneNumberMatcher(text, region, leniency, maxTries)
}

func hasNoMatches(m *phoneNumberMatcher) bool { return !m.hasNext() }

func collectMatches(seq iter.Seq[*PhoneNumberMatch]) []*PhoneNumberMatch {
	var out []*PhoneNumberMatch
	for m := range seq {
		out = append(out, m)
	}
	return out
}

func assertMatchEquals(t *testing.T, expected, actual *PhoneNumberMatch) {
	t.Helper()
	require.NotNil(t, actual)
	assert.Equal(t, expected.RawString(), actual.RawString())
	assert.Equal(t, expected.Start(), actual.Start())
	assert.True(t, proto.Equal(expected.Number(), actual.Number()),
		"number mismatch: %v != %v", expected.Number(), actual.Number())
}

type numberContext struct {
	leadingText  string
	trailingText string
}

type numberTest struct {
	rawString string
	region    string
}

func (nt numberTest) String() string { return nt.rawString + " (" + nt.region + ")" }

// --- tests -------------------------------------------------------------------

func TestContainsMoreThanOneSlashInNationalNumber(t *testing.T) {
	useTestMetadata(t)
	// A date should return true.
	number := &PhoneNumber{CountryCode: proto.Int32(1), CountryCodeSource: ccSource(PhoneNumber_FROM_DEFAULT_COUNTRY)}
	assert.True(t, ContainsMoreThanOneSlashInNationalNumber(number, "1/05/2013"))

	// The country code source thinks it started with a country calling code, but
	// this is not the same as the part before the slash, so it's still true.
	number = &PhoneNumber{CountryCode: proto.Int32(274), CountryCodeSource: ccSource(PhoneNumber_FROM_NUMBER_WITHOUT_PLUS_SIGN)}
	assert.True(t, ContainsMoreThanOneSlashInNationalNumber(number, "27/4/2013"))

	// Now false, because the first slash is after the country calling code.
	number = &PhoneNumber{CountryCode: proto.Int32(49), CountryCodeSource: ccSource(PhoneNumber_FROM_NUMBER_WITH_PLUS_SIGN)}
	assert.False(t, ContainsMoreThanOneSlashInNationalNumber(number, "49/69/2013"))

	number = &PhoneNumber{CountryCode: proto.Int32(49), CountryCodeSource: ccSource(PhoneNumber_FROM_NUMBER_WITHOUT_PLUS_SIGN)}
	assert.False(t, ContainsMoreThanOneSlashInNationalNumber(number, "+49/69/2013"))
	assert.False(t, ContainsMoreThanOneSlashInNationalNumber(number, "+ 49/69/2013"))
	assert.True(t, ContainsMoreThanOneSlashInNationalNumber(number, "+ 49/69/20/13"))

	// Here the first group is not assumed to be the country calling code, even
	// though it is the same as it, so this should return true.
	number = &PhoneNumber{CountryCode: proto.Int32(49), CountryCodeSource: ccSource(PhoneNumber_FROM_DEFAULT_COUNTRY)}
	assert.True(t, ContainsMoreThanOneSlashInNationalNumber(number, "49/69/2013"))
}

func TestFindNationalNumber(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "033316005", regionCode.NZ)
	doTestFindInContext(t, "03-331 6005", regionCode.NZ)
	doTestFindInContext(t, "03 331 6005", regionCode.NZ)
	doTestFindInContext(t, "0064 3 331 6005", regionCode.NZ)
	doTestFindInContext(t, "01164 3 331 6005", regionCode.US)
	doTestFindInContext(t, "+64 3 331 6005", regionCode.US)
	doTestFindInContext(t, "64(0)64123456", regionCode.NZ)
	doTestFindInContext(t, "0123/456789", regionCode.PL)
	doTestFindInContext(t, "123-456-7890", regionCode.US)
}

func TestFindWithInternationalPrefixes(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "+1 (650) 333-6000", regionCode.NZ)
	doTestFindInContext(t, "1-650-333-6000", regionCode.US)
	doTestFindInContext(t, "0011-650-333-6000", regionCode.SG)
	doTestFindInContext(t, "0081-650-333-6000", regionCode.SG)
	doTestFindInContext(t, "0191-650-333-6000", regionCode.SG)
	doTestFindInContext(t, "0~01-650-333-6000", regionCode.PL)
	doTestFindInContext(t, "++1 (650) 333-6000", regionCode.PL)
	doTestFindInContext(t, "＋1 (650) 333-6000", regionCode.SG)
	doTestFindInContext(t, "＋１　（６５０）"+
		"　３３３－６０００", regionCode.SG)
}

func TestFindWithLeadingZero(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "+39 02-36618 300", regionCode.NZ)
	doTestFindInContext(t, "02-36618 300", regionCode.IT)
	doTestFindInContext(t, "312 345 678", regionCode.IT)
}

func TestFindNationalNumberArgentina(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "+54 9 343 555 1212", regionCode.AR)
	doTestFindInContext(t, "0343 15 555 1212", regionCode.AR)
	doTestFindInContext(t, "+54 9 3715 65 4320", regionCode.AR)
	doTestFindInContext(t, "03715 15 65 4320", regionCode.AR)
	doTestFindInContext(t, "+54 11 3797 0000", regionCode.AR)
	doTestFindInContext(t, "011 3797 0000", regionCode.AR)
	doTestFindInContext(t, "+54 3715 65 4321", regionCode.AR)
	doTestFindInContext(t, "03715 65 4321", regionCode.AR)
	doTestFindInContext(t, "+54 23 1234 0000", regionCode.AR)
	doTestFindInContext(t, "023 1234 0000", regionCode.AR)
}

func TestFindWithXInNumber(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "(0xx) 123456789", regionCode.AR)
	doTestFindInContext(t, "(0xx) 123456789 x 1234", regionCode.AR)
	doTestFindInContext(t, "011xx5481429712", regionCode.US)
}

func TestFindNumbersMexico(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "+52 (449)978-0001", regionCode.MX)
	doTestFindInContext(t, "01 (449)978-0001", regionCode.MX)
	doTestFindInContext(t, "(449)978-0001", regionCode.MX)
	doTestFindInContext(t, "+52 1 33 1234-5678", regionCode.MX)
	doTestFindInContext(t, "044 (33) 1234-5678", regionCode.MX)
	doTestFindInContext(t, "045 33 1234-5678", regionCode.MX)
}

func TestFindNumbersWithPlusWithNoRegion(t *testing.T) {
	useTestMetadata(t)
	// ZZ is allowed only if the number starts with a '+' - then the country code
	// can be calculated.
	doTestFindInContext(t, "+64 3 331 6005", regionCode.ZZ)
	// Empty (Java null) is also allowed for the region code in these cases.
	doTestFindInContext(t, "+64 3 331 6005", "")
}

func TestFindExtensions(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "03 331 6005 ext 3456", regionCode.NZ)
	doTestFindInContext(t, "03-3316005x3456", regionCode.NZ)
	doTestFindInContext(t, "03-3316005 int.3456", regionCode.NZ)
	doTestFindInContext(t, "03 3316005 #3456", regionCode.NZ)
	doTestFindInContext(t, "0~0 1800 7493 524", regionCode.PL)
	doTestFindInContext(t, "(1800) 7493.524", regionCode.US)
	doTestFindInContext(t, "0~0 1800 7493 524 ~1234", regionCode.PL)
	doTestFindInContext(t, "+44 2034567890x456", regionCode.NZ)
	doTestFindInContext(t, "+44 2034567890x456", regionCode.GB)
	doTestFindInContext(t, "+44 2034567890 x456", regionCode.GB)
	doTestFindInContext(t, "+44 2034567890 X456", regionCode.GB)
	doTestFindInContext(t, "+44 2034567890 X 456", regionCode.GB)
	doTestFindInContext(t, "+44 2034567890 X  456", regionCode.GB)
	doTestFindInContext(t, "+44 2034567890  X 456", regionCode.GB)
	doTestFindInContext(t, "(800) 901-3355 x 7246433", regionCode.US)
	doTestFindInContext(t, "(800) 901-3355 , ext 7246433", regionCode.US)
	doTestFindInContext(t, "(800) 901-3355 ,extension 7246433", regionCode.US)
	doTestFindInContext(t, "(800) 901-3355 ,x 7246433", regionCode.US)
	doTestFindInContext(t, "(800) 901-3355 ext: 7246433", regionCode.US)
}

func TestFindInterspersedWithSpace(t *testing.T) {
	useTestMetadata(t)
	doTestFindInContext(t, "0 3   3 3 1   6 0 0 5", regionCode.NZ)
}

func TestIntermediateParsePositions(t *testing.T) {
	useTestMetadata(t)
	text := "Call 033316005  or 032316005!"
	//       |    |    |    |    |    |
	//       0    5   10   15   20   25
	for i := 0; i <= 5; i++ {
		assertEqualRange(t, text, i, 5, 14)
	}
	// 7 and 8 digits in a row are still parsed as number.
	assertEqualRange(t, text, 6, 6, 14)
	assertEqualRange(t, text, 7, 7, 14)
	// Anything smaller is skipped to the second instance.
	for i := 8; i <= 19; i++ {
		assertEqualRange(t, text, i, 19, 28)
	}
}

func TestFourMatchesInARow(t *testing.T) {
	useTestMetadata(t)
	number1, number2 := "415-666-7777", "800-443-1223"
	number3, number4 := "212-443-1223", "650-443-1223"
	text := number1 + " - " + number2 + " - " + number3 + " - " + number4

	m := findMatcher(text, regionCode.US)
	assertMatchProperties(t, m.next(), text, number1, regionCode.US)
	assertMatchProperties(t, m.next(), text, number2, regionCode.US)
	assertMatchProperties(t, m.next(), text, number3, regionCode.US)
	assertMatchProperties(t, m.next(), text, number4, regionCode.US)
}

func TestMatchesFoundWithMultipleSpaces(t *testing.T) {
	useTestMetadata(t)
	number1, number2 := "(415) 666-7777", "(800) 443-1223"
	text := number1 + " " + number2

	m := findMatcher(text, regionCode.US)
	assertMatchProperties(t, m.next(), text, number1, regionCode.US)
	assertMatchProperties(t, m.next(), text, number2, regionCode.US)
}

func TestMatchWithSurroundingZipcodes(t *testing.T) {
	useTestMetadata(t)
	number := "415-666-7777"
	zipPreceding := "My address is CA 34215 - " + number + " is my number."
	m := findMatcher(zipPreceding, regionCode.US)
	assertMatchProperties(t, m.next(), zipPreceding, number, regionCode.US)

	// Now the phone number has spaces in it. It should still be found.
	number = "(415) 666 7777"
	zipFollowing := "My number is " + number + ". 34215 is my zip-code."
	m = findMatcher(zipFollowing, regionCode.US)
	assertMatchProperties(t, m.next(), zipFollowing, number, regionCode.US)
}

func TestIsLatinLetter(t *testing.T) {
	assert.True(t, isLatinLetter('c'))
	assert.True(t, isLatinLetter('C'))
	assert.True(t, isLatinLetter('É'))
	assert.True(t, isLatinLetter('́')) // Combining acute accent
	// Punctuation, digits and white-space are not "latin letters".
	assert.False(t, isLatinLetter(':'))
	assert.False(t, isLatinLetter('5'))
	assert.False(t, isLatinLetter('-'))
	assert.False(t, isLatinLetter('.'))
	assert.False(t, isLatinLetter(' '))
	assert.False(t, isLatinLetter('我')) // Chinese character
	assert.False(t, isLatinLetter('の')) // Hiragana letter no
}

func TestMatchesWithSurroundingLatinChars(t *testing.T) {
	useTestMetadata(t)
	contexts := []numberContext{
		{"abc", "def"},
		{"abc", ""},
		{"", "def"},
		{"É", ""},  // Latin capital E with acute accent.
		{"é", ""}, // e with acute accent decomposed (combining mark).
	}
	// Not valid when surrounded by Latin characters, but should be possible.
	findMatchesInContexts(t, contexts, false, true)
}

func TestMoneyNotSeenAsPhoneNumber(t *testing.T) {
	useTestMetadata(t)
	contexts := []numberContext{
		{"$", ""},
		{"", "$"},
		{"£", ""}, // Pound sign
		{"¥", ""}, // Yen sign
	}
	findMatchesInContexts(t, contexts, false, true)
}

func TestPercentageNotSeenAsPhoneNumber(t *testing.T) {
	useTestMetadata(t)
	contexts := []numberContext{{"", "%"}}
	findMatchesInContexts(t, contexts, false, true)
}

func TestPhoneNumberWithLeadingOrTrailingMoneyMatches(t *testing.T) {
	useTestMetadata(t)
	// Because of the space after the 20 (or before the 100) these dollar amounts
	// should not stop the actual number from being found.
	contexts := []numberContext{{"$20 ", ""}, {"", " 100$"}}
	findMatchesInContexts(t, contexts, true, true)
}

func TestMatchesWithSurroundingLatinCharsAndLeadingPunctuation(t *testing.T) {
	useTestMetadata(t)
	// Contexts with trailing characters. Leading characters are okay here since
	// the numbers start with punctuation, but trailing characters are not.
	possibleOnlyContexts := []numberContext{
		{"abc", "def"},
		{"", "def"},
		{"", "É"},
	}
	numberWithPlus := "+14156667777"
	numberWithBrackets := "(415)6667777"
	findMatchesInContextsFor(t, possibleOnlyContexts, false, true, regionCode.US, numberWithPlus)
	findMatchesInContextsFor(t, possibleOnlyContexts, false, true, regionCode.US, numberWithBrackets)

	validContexts := []numberContext{
		{"abc", ""},
		{"É", ""},
		{"É", "."},    // Trailing punctuation.
		{"É", " def"}, // Trailing white-space.
	}
	findMatchesInContextsFor(t, validContexts, true, true, regionCode.US, numberWithPlus)
	findMatchesInContextsFor(t, validContexts, true, true, regionCode.US, numberWithBrackets)
}

func TestMatchesWithSurroundingChineseChars(t *testing.T) {
	useTestMetadata(t)
	validContexts := []numberContext{
		{"我的电话号码是", ""},
		{"", "是我的电话号码"},
		{"请拨打", "我在明天"},
	}
	findMatchesInContexts(t, validContexts, true, true)
}

func TestMatchesWithSurroundingPunctuation(t *testing.T) {
	useTestMetadata(t)
	validContexts := []numberContext{
		{"My number-", ""},          // At end of text.
		{"", ".Nice day."},          // At start of text.
		{"Tel:", "."},               // Punctuation surrounds number.
		{"Tel: ", " on Saturdays."}, // White-space is also fine.
	}
	findMatchesInContexts(t, validContexts, true, true)
}

func TestMatchesMultiplePhoneNumbersSeparatedByPhoneNumberPunctuation(t *testing.T) {
	useTestMetadata(t)
	text := "Call 650-253-4561 -- 455-234-3451"
	region := regionCode.US

	number1 := &PhoneNumber{CountryCode: proto.Int32(int32(GetCountryCodeForRegion(region))), NationalNumber: proto.Uint64(6502534561)}
	match1 := newPhoneNumberMatch(5, "650-253-4561", number1)

	number2 := &PhoneNumber{CountryCode: proto.Int32(int32(GetCountryCodeForRegion(region))), NationalNumber: proto.Uint64(4552343451)}
	match2 := newPhoneNumberMatch(21, "455-234-3451", number2)

	m := findMatcher(text, region)
	assertMatchEquals(t, match1, m.next())
	assertMatchEquals(t, match2, m.next())
}

func TestDoesNotMatchMultiplePhoneNumbersSeparatedWithNoWhiteSpace(t *testing.T) {
	useTestMetadata(t)
	// No white-space found between numbers - neither is found.
	text := "Call 650-253-4561--455-234-3451"
	assert.True(t, hasNoMatches(findMatcher(text, regionCode.US)))
}

// Strings with number-like things that shouldn't be found under any level.
var impossibleCases = []numberTest{
	{"12345", regionCode.US},
	{"23456789", regionCode.US},
	{"234567890112", regionCode.US},
	{"650+253+1234", regionCode.US},
	{"3/10/1984", regionCode.CA},
	{"03/27/2011", regionCode.US},
	{"31/8/2011", regionCode.US},
	{"1/12/2011", regionCode.US},
	{"10/12/82", regionCode.DE},
	{"650x2531234", regionCode.US},
	{"2012-01-02 08:00", regionCode.US},
	{"2012/01/02 08:00", regionCode.US},
	{"20120102 08:00", regionCode.US},
	{"2014-04-12 04:04 PM", regionCode.US},
	{"2014-04-12 &nbsp;04:04 PM", regionCode.US},
	{"2014-04-12 &nbsp;04:04 PM", regionCode.US},
	{"2014-04-12  04:04 PM", regionCode.US},
}

// Strings with number-like things that should only be found under "possible".
var possibleOnlyCases = []numberTest{
	{"7121115678", regionCode.US}, // US numbers cannot start with 7 in test metadata.
	{"1650 x 253 - 1234", regionCode.US},
	{"650 x 253 - 1234", regionCode.US},
	{"6502531x234", regionCode.US},
	{"(20) 3346 1234", regionCode.GB}, // Non-optional NP omitted
}

// Strings that should only be found up to and including the "valid" level.
var validCases = []numberTest{
	{"65 02 53 00 00", regionCode.US},
	{"6502 538365", regionCode.US},
	{"650//253-1234", regionCode.US}, // 2 slashes are illegal at higher levels
	{"650/253/1234", regionCode.US},
	{"9002309. 158", regionCode.US},
	{"12 7/8 - 14 12/34 - 5", regionCode.US},
	{"12.1 - 23.71 - 23.45", regionCode.US},
	{"800 234 1 111x1111", regionCode.US},
	{"1979-2011 100", regionCode.US},
	{"+494949-4-94", regionCode.DE}, // National number in wrong format
	{"４１５６６６６-７７７", regionCode.US},
	{"2012-0102 08", regionCode.US}, // Very strange formatting.
	{"2012-01-02 08", regionCode.US},
	{"1800-1-0-10 22", regionCode.AU}, // Breakdown assistance number.
	{"030-3-2 23 12 34", regionCode.DE},
	{"03 0 -3 2 23 12 34", regionCode.DE},
	{"(0)3 0 -3 2 23 12 34", regionCode.DE},
	{"0 3 0 -3 2 23 12 34", regionCode.DE},
	{"+52 332 123 23 23", regionCode.MX}, // Fits an alternate pattern, but leading digits don't match.
}

// Strings that should only be found up to and including "strict_grouping".
var strictGroupingCases = []numberTest{
	{"(415) 6667777", regionCode.US},
	{"415-6667777", regionCode.US},
	// Found by strict grouping but not exact grouping, as the last two groups are
	// formatted together as a block.
	{"0800-2491234", regionCode.DE},
	// Almost matches an alternate format (last two groups squashed together).
	{"0900-1 123123", regionCode.DE},
	{"(0)900-1 123123", regionCode.DE},
	{"0 900-1 123123", regionCode.DE},
	// NDC also found as part of the country calling code; shouldn't ruin grouping.
	{"+33 3 34 2312", regionCode.FR},
}

// Strings with number-like things that should be found at all levels.
var exactGroupingCases = []numberTest{
	{"４１５６６６７７７７", regionCode.US},
	{"４１５-６６６-７７７７", regionCode.US},
	{"4156667777", regionCode.US},
	{"4156667777 x 123", regionCode.US},
	{"415-666-7777", regionCode.US},
	{"415/666-7777", regionCode.US},
	{"415-666-7777 ext. 503", regionCode.US},
	{"1 415 666 7777 x 123", regionCode.US},
	{"+1 415-666-7777", regionCode.US},
	{"+494949 49", regionCode.DE},
	{"+49-49-34", regionCode.DE},
	{"+49-4931-49", regionCode.DE},
	{"04931-49", regionCode.DE},   // With National Prefix
	{"+49-494949", regionCode.DE}, // One group with country code
	{"+49-494949 ext. 49", regionCode.DE},
	{"+49494949 ext. 49", regionCode.DE},
	{"0494949", regionCode.DE},
	{"0494949 ext. 49", regionCode.DE},
	{"01 (33) 3461 2234", regionCode.MX}, // Optional NP present
	{"(33) 3461 2234", regionCode.MX},    // Optional NP omitted
	{"1800-10-10 22", regionCode.AU},     // Breakdown assistance number.
	// Matches an alternate format exactly.
	{"0900-1 123 123", regionCode.DE},
	{"(0)900-1 123 123", regionCode.DE},
	{"0 900-1 123 123", regionCode.DE},
	{"+33 3 34 23 12", regionCode.FR},
}

func TestMatchesWithPossibleLeniency(t *testing.T) {
	useTestMetadata(t)
	var testCases []numberTest
	testCases = append(testCases, strictGroupingCases...)
	testCases = append(testCases, exactGroupingCases...)
	testCases = append(testCases, validCases...)
	testCases = append(testCases, possibleOnlyCases...)
	doTestNumberMatchesForLeniency(t, testCases, POSSIBLE)
}

func TestNonMatchesWithPossibleLeniency(t *testing.T) {
	useTestMetadata(t)
	doTestNumberNonMatchesForLeniency(t, impossibleCases, POSSIBLE)
}

func TestMatchesWithValidLeniency(t *testing.T) {
	useTestMetadata(t)
	var testCases []numberTest
	testCases = append(testCases, strictGroupingCases...)
	testCases = append(testCases, exactGroupingCases...)
	testCases = append(testCases, validCases...)
	doTestNumberMatchesForLeniency(t, testCases, VALID)
}

func TestNonMatchesWithValidLeniency(t *testing.T) {
	useTestMetadata(t)
	var testCases []numberTest
	testCases = append(testCases, impossibleCases...)
	testCases = append(testCases, possibleOnlyCases...)
	doTestNumberNonMatchesForLeniency(t, testCases, VALID)
}

func TestMatchesWithStrictGroupingLeniency(t *testing.T) {
	useTestMetadata(t)
	var testCases []numberTest
	testCases = append(testCases, strictGroupingCases...)
	testCases = append(testCases, exactGroupingCases...)
	doTestNumberMatchesForLeniency(t, testCases, STRICT_GROUPING)
}

func TestNonMatchesWithStrictGroupLeniency(t *testing.T) {
	useTestMetadata(t)
	var testCases []numberTest
	testCases = append(testCases, impossibleCases...)
	testCases = append(testCases, possibleOnlyCases...)
	testCases = append(testCases, validCases...)
	doTestNumberNonMatchesForLeniency(t, testCases, STRICT_GROUPING)
}

func TestMatchesWithExactGroupingLeniency(t *testing.T) {
	useTestMetadata(t)
	doTestNumberMatchesForLeniency(t, exactGroupingCases, EXACT_GROUPING)
}

func TestNonMatchesExactGroupLeniency(t *testing.T) {
	useTestMetadata(t)
	var testCases []numberTest
	testCases = append(testCases, impossibleCases...)
	testCases = append(testCases, possibleOnlyCases...)
	testCases = append(testCases, validCases...)
	testCases = append(testCases, strictGroupingCases...)
	doTestNumberNonMatchesForLeniency(t, testCases, EXACT_GROUPING)
}

func doTestNumberMatchesForLeniency(t *testing.T, testCases []numberTest, leniency Leniency) {
	noMatchFoundCount := 0
	wrongMatchFoundCount := 0
	for _, test := range testCases {
		m := findMatcherLeniency(test.rawString, test.region, leniency, math.MaxInt)
		var match *PhoneNumberMatch
		if m.hasNext() {
			match = m.next()
		}
		if match == nil {
			noMatchFoundCount++
			t.Logf("No match found in %s for leniency: %d", test, leniency)
		} else if test.rawString != match.RawString() {
			wrongMatchFoundCount++
			t.Logf("Found wrong match in test %s. Found %s", test, match.RawString())
		}
	}
	assert.Equal(t, 0, noMatchFoundCount)
	assert.Equal(t, 0, wrongMatchFoundCount)
}

func doTestNumberNonMatchesForLeniency(t *testing.T, testCases []numberTest, leniency Leniency) {
	matchFoundCount := 0
	for _, test := range testCases {
		m := findMatcherLeniency(test.rawString, test.region, leniency, math.MaxInt)
		if m.hasNext() {
			matchFoundCount++
			t.Logf("Match found in %s for leniency: %d", test, leniency)
		}
	}
	assert.Equal(t, 0, matchFoundCount)
}

// findMatchesInContexts tests the contexts with a default number/region.
func findMatchesInContexts(t *testing.T, contexts []numberContext, isValid, isPossible bool) {
	findMatchesInContextsFor(t, contexts, isValid, isPossible, regionCode.US, "415-666-7777")
}

func findMatchesInContextsFor(t *testing.T, contexts []numberContext, isValid, isPossible bool, region, number string) {
	if isValid {
		doTestInContext(t, number, region, contexts, VALID)
	} else {
		for _, context := range contexts {
			text := context.leadingText + number + context.trailingText
			assert.True(t, hasNoMatches(findMatcher(text, region)), "Should not have found a number in %q", text)
		}
	}
	if isPossible {
		doTestInContext(t, number, region, contexts, POSSIBLE)
	} else {
		for _, context := range contexts {
			text := context.leadingText + number + context.trailingText
			assert.True(t, hasNoMatches(findMatcherLeniency(text, region, POSSIBLE, math.MaxInt)),
				"Should not have found a number in %q", text)
		}
	}
}

func TestNonMatchingBracketsAreInvalid(t *testing.T) {
	useTestMetadata(t)
	// The digits up to the ", " form a valid US number, but shouldn't be matched
	// since there's a non-matching bracket present.
	assert.True(t, hasNoMatches(findMatcher("80.585 [79.964, 81.191]", regionCode.US)))
	// The trailing "]" is thrown away before parsing, so the number doesn't have
	// matching brackets.
	assert.True(t, hasNoMatches(findMatcher("80.585 [79.964]", regionCode.US)))
	assert.True(t, hasNoMatches(findMatcher("80.585 ((79.964)", regionCode.US)))
	// Too many sets of brackets to be valid.
	assert.True(t, hasNoMatches(findMatcher("(80).(585) (79).(9)64", regionCode.US)))
}

func TestNoMatchIfRegionIsNull(t *testing.T) {
	useTestMetadata(t)
	// Fail on non-international prefix if region code is empty (Java null).
	assert.True(t, hasNoMatches(findMatcher("Random text body - number is 0331 6005, see you there", "")))
}

func TestNoMatchInEmptyString(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, hasNoMatches(findMatcher("", regionCode.US)))
	assert.True(t, hasNoMatches(findMatcher("  ", regionCode.US)))
}

func TestNoMatchIfNoNumber(t *testing.T) {
	useTestMetadata(t)
	assert.True(t, hasNoMatches(findMatcher("Random text body - number is foobar, see you there", regionCode.US)))
}

func TestSequences(t *testing.T) {
	useTestMetadata(t)
	// Test multiple occurrences.
	text := "Call 033316005  or 032316005!"
	region := regionCode.NZ

	number1 := &PhoneNumber{CountryCode: proto.Int32(int32(GetCountryCodeForRegion(region))), NationalNumber: proto.Uint64(33316005)}
	match1 := newPhoneNumberMatch(5, "033316005", number1)

	number2 := &PhoneNumber{CountryCode: proto.Int32(int32(GetCountryCodeForRegion(region))), NationalNumber: proto.Uint64(32316005)}
	match2 := newPhoneNumberMatch(19, "032316005", number2)

	m := findMatcherLeniency(text, region, POSSIBLE, math.MaxInt)
	assertMatchEquals(t, match1, m.next())
	assertMatchEquals(t, match2, m.next())
}

func TestNullInput(t *testing.T) {
	useTestMetadata(t)
	// Java passes null text; the Go API takes a string, so we use "".
	assert.True(t, hasNoMatches(findMatcher("", regionCode.US)))
	assert.True(t, hasNoMatches(findMatcher("", "")))
}

func TestMaxMatches(t *testing.T) {
	useTestMetadata(t)
	// Set up text with 100 valid phone numbers.
	var numbers strings.Builder
	for i := 0; i < 100; i++ {
		numbers.WriteString("My info: 415-666-7777,")
	}

	// Matches all 100. Max only applies to failed cases.
	number, err := Parse("+14156667777", "")
	require.NoError(t, err)

	matches := collectMatches(FindNumbersWithLeniency(numbers.String(), regionCode.US, VALID, 10))
	assert.Len(t, matches, 100)
	for _, match := range matches {
		assert.True(t, proto.Equal(number, match.Number()))
	}
}

func TestMaxMatchesInvalid(t *testing.T) {
	useTestMetadata(t)
	// 10 invalid phone numbers followed by 100 valid.
	var numbers strings.Builder
	for i := 0; i < 10; i++ {
		numbers.WriteString("My address 949-8945-0")
	}
	for i := 0; i < 100; i++ {
		numbers.WriteString("My info: 415-666-7777,")
	}
	m := findMatcherLeniency(numbers.String(), regionCode.US, VALID, 10)
	assert.False(t, m.hasNext())
}

func TestMaxMatchesMixed(t *testing.T) {
	useTestMetadata(t)
	// 100 valid numbers inside an invalid number.
	var numbers strings.Builder
	for i := 0; i < 100; i++ {
		numbers.WriteString("My info: 415-666-7777 123 fake street")
	}

	number, err := Parse("+14156667777", "")
	require.NoError(t, err)

	// Only matches the first 10 despite there being 100 numbers due to max matches.
	matches := collectMatches(FindNumbersWithLeniency(numbers.String(), regionCode.US, VALID, 10))
	assert.Len(t, matches, 10)
	for _, match := range matches {
		assert.True(t, proto.Equal(number, match.Number()))
	}
}

func TestNonPlusPrefixedNumbersNotFoundForInvalidRegion(t *testing.T) {
	useTestMetadata(t)
	// Does not start with a "+", we won't match it.
	m := findMatcher("1 456 764 156", regionCode.ZZ)
	assert.False(t, m.hasNext())
	assert.Nil(t, m.next()) // Java throws NoSuchElementException; Go returns nil.
	assert.False(t, m.hasNext())
}

func TestEmptyIteration(t *testing.T) {
	useTestMetadata(t)
	m := findMatcher("", regionCode.ZZ)
	assert.False(t, m.hasNext())
	assert.False(t, m.hasNext())
	assert.Nil(t, m.next())
	assert.False(t, m.hasNext())
}

func TestSingleIteration(t *testing.T) {
	useTestMetadata(t)
	// With hasNext() -> next().
	m := findMatcher("+14156667777", regionCode.ZZ)
	// Double hasNext() to ensure it does not advance.
	assert.True(t, m.hasNext())
	assert.True(t, m.hasNext())
	assert.NotNil(t, m.next())
	assert.False(t, m.hasNext())
	assert.Nil(t, m.next())
	assert.False(t, m.hasNext())

	// With next() only.
	m = findMatcher("+14156667777", regionCode.ZZ)
	assert.NotNil(t, m.next())
	assert.Nil(t, m.next())
}

func TestDoubleIteration(t *testing.T) {
	useTestMetadata(t)
	// With hasNext() -> next().
	m := findMatcher("+14156667777 foobar +14156667777 ", regionCode.ZZ)
	assert.True(t, m.hasNext())
	assert.True(t, m.hasNext())
	assert.NotNil(t, m.next())
	assert.True(t, m.hasNext())
	assert.True(t, m.hasNext())
	assert.NotNil(t, m.next())
	assert.False(t, m.hasNext())
	assert.Nil(t, m.next())
	assert.False(t, m.hasNext())

	// With next() only.
	m = findMatcher("+14156667777 foobar +14156667777 ", regionCode.ZZ)
	assert.NotNil(t, m.next())
	assert.NotNil(t, m.next())
	assert.Nil(t, m.next())
}

// --- shared assertion helpers ------------------------------------------------

func assertMatchProperties(t *testing.T, match *PhoneNumberMatch, text, number, region string) {
	t.Helper()
	expectedResult, err := Parse(number, region)
	require.NoError(t, err)
	require.NotNil(t, match, "Did not find a number in '%s'; expected %s", text, number)
	assert.True(t, proto.Equal(expectedResult, match.Number()), "number mismatch in %q", text)
	assert.Equal(t, number, match.RawString())
}

func assertEqualRange(t *testing.T, text string, index, start, end int) {
	t.Helper()
	sub := text[index:]
	m := findMatcherLeniency(sub, regionCode.NZ, POSSIBLE, math.MaxInt)
	require.True(t, m.hasNext())
	match := m.next()
	assert.Equal(t, start-index, match.Start())
	assert.Equal(t, end-index, match.End())
	assert.Equal(t, sub[match.Start():match.End()], match.RawString())
}

func doTestFindInContext(t *testing.T, number, defaultCountry string) {
	findPossibleInContext(t, number, defaultCountry)
	parsed, err := Parse(number, defaultCountry)
	require.NoError(t, err)
	if IsValidNumber(parsed) {
		findValidInContext(t, number, defaultCountry)
	}
}

func findPossibleInContext(t *testing.T, number, defaultCountry string) {
	contextPairs := []numberContext{
		{"", ""},                             // no context
		{"   ", "\t"},                        // whitespace only
		{"Hello ", ""},                       // no context at end
		{"", " to call me!"},                 // no context at start
		{"Hi there, call ", " to reach me!"}, // no context at start
		{"Hi there, call ", ", or don't"},    // with commas
		{"Hi call", ""},
		{"", "forme"},
		{"Hi call", "forme"},
		{"It's cheap! Call ", " before 6:30"}, // with other small numbers
		{"Call ", " or +1800-123-4567!"},      // with a second number later
		{"Call me on June 2 at", ""},          // with a Month-Day date
		{"As quoted by Alfonso 12-15 (2009), you may call me at ", ""},
		{"As quoted by Alfonso et al. 12-15 (2009), you may call me at ", ""},
		{"As I said on 03/10/2011, you may call me at ", ""},
		{"", ", 45 days a year"},
		{"", ";x 7246433"},
		{"Call ", "/x12 more"},
	}
	doTestInContext(t, number, defaultCountry, contextPairs, POSSIBLE)
}

func findValidInContext(t *testing.T, number, defaultCountry string) {
	contextPairs := []numberContext{
		{"It's only 9.99! Call ", " to buy"}, // with other small numbers
		{"Call me on 21.6.1984 at ", ""},     // Day.Month.Year date
		{"Call me on 06/21 at ", ""},         // Month/Day date
		{"Call me on 21.6. at ", ""},         // Day.Month date
		{"Call me on 06/21/84 at ", ""},      // Month/Day/Year date
	}
	doTestInContext(t, number, defaultCountry, contextPairs, VALID)
}

func doTestInContext(t *testing.T, number, defaultCountry string, contextPairs []numberContext, leniency Leniency) {
	for _, context := range contextPairs {
		prefix := context.leadingText
		text := prefix + number + context.trailingText

		start := len(prefix)
		end := start + len(number)
		m := findMatcherLeniency(text, defaultCountry, leniency, math.MaxInt)

		var match *PhoneNumberMatch
		if m.hasNext() {
			match = m.next()
		}
		require.NotNil(t, match, "Did not find a number in '%s'; expected '%s'", text, number)

		extracted := text[match.Start():match.End()]
		assert.True(t, start == match.Start() && end == match.End(),
			"Unexpected phone region in '%s'; extracted '%s'", text, extracted)
		assert.Equal(t, number, extracted)
		assert.Equal(t, extracted, match.RawString())

		ensureTermination(t, text, defaultCountry, leniency)
	}
}

// ensureTermination exhaustively searches from each byte index within text to
// test that finding matches always terminates. It exercises the public
// FindNumbersWithLeniency iterator.
func ensureTermination(t *testing.T, text, defaultCountry string, leniency Leniency) {
	for index := 0; index <= len(text); index++ {
		sub := text[index:]
		for range FindNumbersWithLeniency(sub, defaultCountry, leniency, math.MaxInt) {
			// Iterate over all matches; we only care that it terminates.
		}
	}
}
