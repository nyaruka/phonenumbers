package phonenumbers

// Go-specific unit tests with no standalone counterpart in upstream's
// PhoneNumberUtilTest: direct tests of internal helpers (normalizeDigits's
// keep-non-digits branch, setItalianLeadingZerosForPhoneNumber,
// maybeStripExtension, mergeLengths, formattingRuleHasFirstGroupOnly) and the
// RegexCache strictness behaviour (which upstream tests in its own internal/
// module).

import (
	"reflect"
	"testing"

	"github.com/nyaruka/phonenumbers/v2/metadata"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNormalizeDigits(t *testing.T) {
	var tests = []struct {
		input         string
		keepNonDigits bool
		expected      []byte
	}{
		{input: "4445556666", keepNonDigits: false, expected: []byte("4445556666")},
		{input: "(444)5556666", keepNonDigits: false, expected: []byte("4445556666")},
		{input: "(444)555a6666", keepNonDigits: false, expected: []byte("4445556666")},
		{input: "(444)555a6666", keepNonDigits: true, expected: []byte("(444)555a6666")},
	}

	for _, tc := range tests {
		actual := normalizeDigits(tc.input, tc.keepNonDigits)
		assert.Equal(t, string(tc.expected), string(actual), "mismatch for input %s", tc.input)
	}
}

func TestSetItalianLeadinZerosForPhoneNumber(t *testing.T) {
	var tests = []struct {
		num          string
		numLeadZeros int32
		hasLeadZero  bool
	}{
		{
			num:          "00000",
			numLeadZeros: 4,
			hasLeadZero:  true,
		},
		{
			num:          "0123456",
			numLeadZeros: 1,
			hasLeadZero:  true,
		},
		{
			num:          "0023456",
			numLeadZeros: 2,
			hasLeadZero:  true,
		},
		{
			num:          "123456",
			numLeadZeros: 1, // it's the default value
			hasLeadZero:  false,
		},
	}

	for i, test := range tests {
		var pNum = &PhoneNumber{}
		setItalianLeadingZerosForPhoneNumber(test.num, pNum)
		if pNum.GetItalianLeadingZero() != test.hasLeadZero {
			t.Errorf("[test %d:hasLeadZero] %v != %v\n",
				i, pNum.GetItalianLeadingZero(), test.hasLeadZero)
		}
		if pNum.GetNumberOfLeadingZeros() != test.numLeadZeros {
			t.Errorf("[test %d:numLeadZeros] %v != %v\n",
				i, pNum.GetNumberOfLeadingZeros(), test.numLeadZeros)
		}

	}
}

func newPhoneNumber(cc int, natNum uint64) *PhoneNumber {
	p := &PhoneNumber{}
	p.CountryCode = proto.Int32(int32(cc))
	p.NationalNumber = proto.Uint64(natNum)
	return p
}

func TestMaybeStripExtension(t *testing.T) {
	var tests = []struct {
		input     string
		number    uint64
		extension string
		region    string
	}{
		{
			input:     "1234576 ext. 1234",
			number:    1234576,
			extension: "1234",
			region:    "US",
		}, {
			input:     "1234-576",
			number:    1234576,
			extension: "",
			region:    "US",
		}, {
			input:     "1234576-123#",
			number:    1234576,
			extension: "123",
			region:    "US",
		}, {
			input:     "1234576 ext.123#",
			number:    1234576,
			extension: "123",
			region:    "US",
		},
		// Russian extension label "доб"
		{
			input:     "8 (423) 202-25-11, \u0434\u043E\u0431. 100",
			number:    4232022511,
			extension: "100",
			region:    "RU",
		},
		{
			input:     "8 (423) 202-25-11 \u0434\u043E\u0431 100",
			number:    4232022511,
			extension: "100",
			region:    "RU",
		},
		// Russian extension label uppercase "ДОБ"
		{
			input:     "8 (423) 202-25-11, \u0414\u041E\u0411. 100",
			number:    4232022511,
			extension: "100",
			region:    "RU",
		},
		// Auto-dialling with ",,"
		{
			input:     "+12679000000,,123456789012345#",
			number:    2679000000,
			extension: "123456789012345",
			region:    "US",
		},
		// Auto-dialling with ";"
		{
			input:     "+12679000000;123456789012345#",
			number:    2679000000,
			extension: "123456789012345",
			region:    "US",
		},
		// Single comma extension
		{
			input:     "+442034000000,123456789#",
			number:    2034000000,
			extension: "123456789",
			region:    "GB",
		},
		// Explicit label with up to 20 digits
		{
			input:     "03 3316005 xtn:12345678901234567890",
			number:    33316005,
			extension: "12345678901234567890",
			region:    "NZ",
		},
		// RFC3966 with up to 20 digits
		{
			input:     "tel:+6433316005;ext=01234567890123456789",
			number:    33316005,
			extension: "01234567890123456789",
			region:    "NZ",
		},
		// Ambiguous char with up to 9 digits
		{
			input:     "03 3316005 x 123456789",
			number:    33316005,
			extension: "123456789",
			region:    "NZ",
		},
		// Trailing # with up to 6 digits
		{
			input:     "+11234567890 666666#",
			number:    1234567890,
			extension: "666666",
			region:    "US",
		},
		// extensión with accented o
		{
			input:     "(800) 901-3355 ,extensi\u00F3n 7246433",
			number:    8009013355,
			extension: "7246433",
			region:    "US",
		},
		// Full-width extension "ｅｘｔｎ"
		{
			input:     "+442034567890\uFF45\uFF58\uFF54\uFF4E456",
			number:    2034567890,
			extension: "456",
			region:    "GB",
		},
	}

	for i, test := range tests {
		num, _ := Parse(test.input, test.region)
		if num.GetNationalNumber() != test.number {
			t.Errorf("[test %d:num] failed: %v != %v\n", i, num.GetNationalNumber(), test.number)
		}

		if num.GetExtension() != test.extension {
			t.Errorf("[test %d:num] failed: %v != %v\n", i, num.GetExtension(), test.extension)
		}
	}
}

// TestGetSupportedCallingCodes was migrated to the faithful upstream port in
// phonenumberutil_test.go (run against synthetic test metadata).

func TestMergeLengths(t *testing.T) {
	var tests = []struct {
		l1     []int32
		l2     []int32
		merged []int32
	}{
		{
			[]int32{1, 5},
			[]int32{2, 3, 4},
			[]int32{1, 2, 3, 4, 5},
		},
		{
			[]int32{1},
			[]int32{3, 4},
			[]int32{1, 3, 4},
		},
		{
			[]int32{1, 2, 5},
			[]int32{4},
			[]int32{1, 2, 4, 5},
		},
	}

	for i, tc := range tests {
		merged := mergeLengths(tc.l1, tc.l2)
		if !reflect.DeepEqual(merged, tc.merged) {
			t.Errorf("[test %d:num] failed for l1: %v and l2: %v: %v != %v\n", i, tc.l1, tc.l2, tc.merged, merged)
		}
	}
}

func TestRegexCacheStrict(t *testing.T) {
	const expectedResult = "(41) 3020-3445"
	phoneToTest := &PhoneNumber{
		CountryCode:    proto.Int32(55),
		NationalNumber: proto.Uint64(4130203445),
	}
	firstRunResult := Format(phoneToTest, NATIONAL)
	if expectedResult != firstRunResult {
		t.Errorf("phone number formatting not as expected")
	}
	// This adds value to the regex cache that would break the following lookup if the regex-s
	// in cache were not strict.
	Format(&PhoneNumber{
		CountryCode:    proto.Int32(973),
		NationalNumber: proto.Uint64(17112724),
	}, NATIONAL)
	secondRunResult := Format(phoneToTest, NATIONAL)

	if expectedResult != secondRunResult {
		t.Errorf("phone number formatting not as expected")
	}
}

func TestFormattingRuleHasFirstGroupOnly(t *testing.T) {
	// Verify the fix: formattingRuleHasFirstGroupOnly should do a full match
	assert.True(t, formattingRuleHasFirstGroupOnly("$1"))
	assert.True(t, formattingRuleHasFirstGroupOnly("($1)"))
	assert.True(t, formattingRuleHasFirstGroupOnly(""))
	assert.False(t, formattingRuleHasFirstGroupOnly("0$1"))
	assert.False(t, formattingRuleHasFirstGroupOnly("0($1)"))
	assert.False(t, formattingRuleHasFirstGroupOnly("$1 suffix"))
}

// TestParseNoPanicWithTelAfterPhoneContext is a regression test for
// GHSA-374v-j3m9-frpm: when "tel:" appears after a valid ";phone-context=",
// buildNationalNumberForParsing used to slice with a start index past its end
// and panic. Parse must return an error instead of crashing.
func TestParseNoPanicWithTelAfterPhoneContext(t *testing.T) {
	for _, in := range []string{
		"5;phone-context=+1;tel:x",
		"5;phone-context=+49;foo=tel:",
	} {
		assert.NotPanics(t, func() {
			_, _ = Parse(in, "US")
		}, "input %q", in)
	}
}

// TestGetNationalSignificantNumberClampsLeadingZeros is a regression test for
// GHSA-4h7c-rcfg-mqmw: NumberOfLeadingZeros is an attacker-controllable int32
// when a PhoneNumber crosses a trust boundary (e.g. protobuf from an untrusted
// source), and it was used unbounded as an allocation size. The count must be
// clamped to maxLengthForNSN so a hostile value cannot drive a huge allocation.
func TestGetNationalSignificantNumberClampsLeadingZeros(t *testing.T) {
	hostile := &PhoneNumber{
		CountryCode:          proto.Int32(1),
		NationalNumber:       proto.Uint64(6502530000),
		ItalianLeadingZero:   proto.Bool(true),
		NumberOfLeadingZeros: proto.Int32(1 << 30),
	}
	nsn := GetNationalSignificantNumber(hostile)
	// maxLengthForNSN leading zeros plus the 10-digit national number.
	assert.Len(t, nsn, maxLengthForNSN+len("6502530000"))
}

func BenchmarkLoadMetadata(b *testing.B) {
	for i := 0; i <= b.N; i++ {
		_, _ = metadata.Load()
	}
}
