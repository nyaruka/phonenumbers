package phonenumbers

// Ad-hoc tests against the REAL embedded metadata, inherited from before the
// upstream test suite was adopted. Because they assert on real-world numbering
// data, they break whenever upstream refreshes metadata.
//
// MIGRATION: these are being replaced by faithful ports of upstream's own tests
// (in phonenumberutil_test.go and the _types_/_format_/_parse_ split files), which
// run against the frozen synthetic test metadata and so never go stale. As each
// upstream test method is ported, delete the corresponding ad-hoc test below. The
// goal is for this file to shrink to (near) nothing.
//
// Residue that has no upstream equivalent in PhoneNumberUtilTest and will linger
// until/unless upstream's other test modules are ported: the geocoding/carrier/
// timezone lookups (upstream java/geocoder + java/carrier test modules), the
// RegexCache tests (upstream internal/), and a handful of Go-port-specific
// real-world parsing regressions (e.g. TestBurkinaFaso, TestMexico).

import (
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestConvertAlphaCharactersInNumber(t *testing.T) {
	var tests = []struct {
		input, expected string
	}{
		{input: "1800AWWPOOP", expected: "18002997667"},
		{input: "(800) DAW-ORLD", expected: "(800) 329-6753"},
		{input: "1800-ABC-DEF", expected: "1800-222-333"},
	}

	for _, tc := range tests {
		actual := ConvertAlphaCharactersInNumber(tc.input)
		assert.Equal(t, tc.expected, actual, "mismatch for input %s", tc.input)
	}
}

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

func TestExtractPossibleNumber(t *testing.T) {
	assert.Equal(t, "530) 583-6985 x302", extractPossibleNumber("(530) 583-6985 x302/x2303")) // yes, the leading '(' is missing
}

func TestIsViablePhoneNumber(t *testing.T) {
	var tests = []struct {
		input    string
		isViable bool
	}{
		{input: "4445556666", isViable: true},
		{input: "+441932123456", isViable: true},
		{input: "4930123456", isViable: true},
		{input: "2", isViable: false},
		{input: "helloworld", isViable: false},
	}

	for _, tc := range tests {
		actual := isViablePhoneNumber(tc.input)
		assert.Equal(t, tc.isViable, actual, "mismatch for input %s", tc.input)
	}
}

func TestNormalize(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{input: "4431234567", expected: "4431234567"},
		{input: "443 1234567", expected: "4431234567"},
		{input: "(443)123-4567", expected: "4431234567"},
		{input: "800yoloFOO", expected: "8009656366"},
		{input: "444111a2222", expected: "4441112222"},

		// from libponenumber [java] unit tests
		{input: "034-56&+#2\u00AD34", expected: "03456234"},
		{input: "034-I-am-HUNGRY", expected: "034426486479"},
		{input: "\uFF125\u0665", expected: "255"},
		{input: "\u06F52\u06F0", expected: "520"},
	}

	for _, tc := range tests {
		actual := normalize(tc.input)
		assert.Equal(t, tc.expected, actual, "mismatch for input %s", tc.input)
	}
}

func TestRepeatedParsing(t *testing.T) {
	phoneNumbers := []string{"+917827202781", "+910000000000", "+910800125778", "+917503257232", "+917566482842"}

	number := &PhoneNumber{}
	for _, n := range phoneNumbers {
		num, err := Parse(n, "IN")
		assert.NoError(t, err, "unexpected error for input %s", n)

		err = ParseToNumber(n, "IN", number)
		assert.NoError(t, err, "unexpected error for input %s", n)

		assert.Equal(t, IsValidNumber(num), IsValidNumber(number))
	}
}

func TestIsPossibleNumberWithReason(t *testing.T) {
	var tests = []struct {
		input  string
		region string
		err    error
		valid  ValidationResult
	}{
		{input: "16502530000", region: "US", err: nil, valid: IS_POSSIBLE},
		{input: "2530000", region: "US", err: nil, valid: IS_POSSIBLE_LOCAL_ONLY},
		{input: "65025300001", region: "US", err: nil, valid: TOO_LONG},
		{input: "2530000", region: "", err: ErrInvalidCountryCode, valid: IS_POSSIBLE_LOCAL_ONLY},
		{input: "253000", region: "US", err: nil, valid: TOO_SHORT},
		{input: "1234567890", region: "SG", err: nil, valid: IS_POSSIBLE},
		{input: "800123456789", region: "US", err: nil, valid: TOO_LONG},
		{input: "+1456723456", region: "US", err: nil, valid: TOO_SHORT},
		{input: "6041234567", region: "US", err: nil, valid: IS_POSSIBLE},
		{input: "+2250749195919", region: "CI", err: nil, valid: IS_POSSIBLE},
	}

	for _, tc := range tests {
		num, err := Parse(tc.input, tc.region)

		if tc.err != nil {
			assert.EqualError(t, err, tc.err.Error(), "error mismatch for input %s", tc.input)
		} else {
			assert.NoError(t, err, "unexpected error for input %s", tc.input)
			assert.Equal(t, tc.valid, IsPossibleNumberWithReason(num), "mismatch for input %s", tc.input)
		}
	}
}

func TestTruncateTooLongNumber(t *testing.T) {
	var tests = []struct {
		country int
		input   uint64
		res     bool
		output  uint64
	}{
		{country: 1, input: 80055501234, res: true, output: 8005550123},
		{country: 1, input: 8005550123, res: true, output: 8005550123},
		{country: 1, input: 800555012, res: false, output: 800555012},
	}

	for _, tc := range tests {
		num := &PhoneNumber{}
		num.CountryCode = proto.Int32(int32(tc.country))
		num.NationalNumber = proto.Uint64(tc.input)
		res := TruncateTooLongNumber(num)

		assert.Equal(t, tc.res, res, "res mismatch for input %d", tc.input)
		assert.Equal(t, tc.output, *num.NationalNumber, "output mismatch for input %d", tc.input)
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

func TestGetMetadata(t *testing.T) {
	var tests = []struct {
		name       string
		id         string
		cc         int32
		i18nPref   string
		natPref    string
		numFmtSize int
	}{
		{
			name:       "US",
			id:         "US",
			cc:         1,
			i18nPref:   "011",
			natPref:    "1",
			numFmtSize: 3,
		}, {
			name:       "DE",
			id:         "DE",
			cc:         49,
			i18nPref:   "00",
			natPref:    "0",
			numFmtSize: 19,
		}, {
			name:       "AR",
			id:         "AR",
			cc:         54,
			i18nPref:   "00",
			natPref:    "0",
			numFmtSize: 12,
		},
	}
	for i, test := range tests {
		meta := getMetadataForRegion(test.name)
		if meta.GetId() != test.name {
			t.Errorf("[test %d:name] %s != %s\n", i, meta.GetId(), test.name)
		}
		if meta.GetCountryCode() != test.cc {
			t.Errorf("[test %d:cc] %d != %d\n", i, meta.GetCountryCode(), test.cc)
		}
		if meta.GetInternationalPrefix() != test.i18nPref {
			t.Errorf("[test %d:i18nPref] %s != %s\n",
				i, meta.GetInternationalPrefix(), test.i18nPref)
		}
		if meta.GetNationalPrefix() != test.natPref {
			t.Errorf("[test %d:natPref] %s != %s\n",
				i, meta.GetNationalPrefix(), test.natPref)
		}
		if len(meta.GetNumberFormat()) != test.numFmtSize {
			t.Errorf("[test %d:numFmt] %d (%s) != %d\n",
				i, len(meta.GetNumberFormat()), meta.GetNumberFormat(), test.numFmtSize)
		}
	}
}

func TestNormalizeDigitsOnly(t *testing.T) {
	if NormalizeDigitsOnly("034-56&+a#234") != "03456234" {
		t.Errorf("didn't fully normalize digits only")
	}
}

func TestNormalizeDiallableCharsOnly(t *testing.T) {
	// '#' is a diallable character and should be preserved
	assert.Equal(t, "03*456+#234", normalizeDiallableCharsOnly("03*4-56&+a#234"))
}

type testCase struct {
	num           string
	parseRegion   string
	expectedE164  string
	validRegion   string
	isValid       bool
	isValidRegion bool
}

type timeZonesTestCases struct {
	num               string
	expectedTimeZones []string
}

func runTestBatch(t *testing.T, tests []testCase) {
	for _, test := range tests {
		n, err := Parse(test.num, test.parseRegion)
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}

		if IsValidNumber(n) != test.isValid {
			t.Errorf("Number %s: validity mismatch: expected %t got %t.", test.num, test.isValid, !test.isValid)
		}

		if IsValidNumberForRegion(n, test.validRegion) != test.isValidRegion {
			t.Errorf("Number %s: region validity mismatch: expected %t got %t.", test.num, test.isValidRegion, !test.isValidRegion)
		}

		s := Format(n, E164)
		if s != test.expectedE164 {
			t.Errorf("Expected '%s', got '%s'", test.expectedE164, s)
		}
	}
}

func TestItalianLeadingZeroes(t *testing.T) {

	tests := []testCase{
		{
			num:           "0491 570 156",
			parseRegion:   "AU",
			expectedE164:  "+61491570156",
			validRegion:   "AU",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "02 5550 1234",
			parseRegion:   "AU",
			expectedE164:  "+61255501234",
			validRegion:   "AU",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "+39.0399123456",
			parseRegion:   "IT",
			expectedE164:  "+390399123456",
			validRegion:   "IT",
			isValid:       true,
			isValidRegion: true,
		},
	}

	runTestBatch(t, tests)
}

func TestARNumberTransformRule(t *testing.T) {
	tests := []testCase{
		{
			num:           "+541151123456",
			parseRegion:   "AR",
			expectedE164:  "+541151123456",
			validRegion:   "AR",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "+540111561234567",
			parseRegion:   "AR",
			expectedE164:  "+5491161234567",
			validRegion:   "AR",
			isValid:       true,
			isValidRegion: true,
		},
	}

	runTestBatch(t, tests)
}

func TestLeadingOne(t *testing.T) {
	tests := []testCase{
		{
			num:           "15167706076",
			parseRegion:   "US",
			expectedE164:  "+15167706076",
			validRegion:   "US",
			isValid:       true,
			isValidRegion: true,
		},
	}

	runTestBatch(t, tests)
}

func TestNewIndianPhones(t *testing.T) {
	tests := []testCase{
		{
			num:           "7999999543",
			parseRegion:   "IN",
			expectedE164:  "+917999999543",
			validRegion:   "IN",
			isValid:       true,
			isValidRegion: true,
		},
	}

	runTestBatch(t, tests)
}

func TestBurkinaFaso(t *testing.T) {
	tests := []testCase{
		{
			num:           "+22658125926",
			parseRegion:   "",
			expectedE164:  "+22658125926",
			validRegion:   "BF",
			isValid:       true,
			isValidRegion: true,
		},
	}

	runTestBatch(t, tests)
}

// see https://groups.google.com/forum/#!topic/libphonenumber-discuss/pecTIo_HpVE
//
// by official decision from the Federal Telecommunication Institute (IFT)(https://www.ift.org.mx/),
// dialing telephone numbers from across Mexico will change to only 10 digits, among other decisions, take note:
// Prefix 01 is eliminated for national calls and non-geographic numbers (880 and 900)
func TestMexico(t *testing.T) {
	tests := []testCase{
		{
			num:           "664 899 1010", // 664 (area code): local 7 digits
			parseRegion:   "MX",
			expectedE164:  "+526648991010",
			validRegion:   "MX",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "55 8912 4785", // 55 (area code): local 8 digits
			parseRegion:   "MX",
			expectedE164:  "+525589124785",
			validRegion:   "MX",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "800 123 2222", // Non-geographic 800
			parseRegion:   "MX",
			expectedE164:  "+528001232222",
			validRegion:   "MX",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "900 433 1234", // Non-geographic 900
			parseRegion:   "MX",
			expectedE164:  "+529004331234",
			validRegion:   "MX",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "+52 664 899 1010", // Long distance international incoming call to cell phone
			parseRegion:   "",
			expectedE164:  "+526648991010",
			validRegion:   "MX",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "228 234 5687", // National Long Distance
			parseRegion:   "MX",
			expectedE164:  "+522282345687",
			validRegion:   "MX",
			isValid:       true,
			isValidRegion: true,
		},
		{
			num:           "+5214424123123", // too long
			parseRegion:   "MX",
			expectedE164:  "+5214424123123",
			validRegion:   "MX",
			isValid:       false,
			isValidRegion: false,
		},
	}

	runTestBatch(t, tests)
}

func TestParsing(t *testing.T) {
	tests := []struct {
		number   string
		country  string
		expected string
	}{
		{"746557816", "LK", "+94746557816"},
		{"0788383383", "RW", "+250788383383"},
		{"+250788383383 ", "KE", "+250788383383"},
		{"+250788383383", "", "+250788383383"},
		{"(917)992-5253", "US", "+19179925253"},
		{"+62877747666", "", "+62877747666"},
		{"0877747666", "ID", "+62877747666"},
		{"07531669965", "GB", "+447531669965"},
		{"447531669965", "GB", "+447531669965"},
		{"+22658125926", "", "+22658125926"},
		{"+2203693200", "", "+2203693200"},
		{"0877747666", "ID", "+62877747666"},
		{"62816640000", "ID", "+6262816640000"},
		{"2349090000001", "NG", "+2349090000001"},
		{"6282240080000", "ID", "+6282240080000"},
	}

	for _, tc := range tests {
		num, err := Parse(tc.number, tc.country)
		if err != nil {
			t.Errorf("Error parsing number: %s: %s", tc.number, err)
		} else {
			formatted := Format(num, E164)
			if formatted != tc.expected {
				t.Errorf("Error parsing number '%s:%s', got %s when expecting %s", tc.number, tc.country, formatted, tc.expected)
			}
		}
	}
}

func TestGetTimeZonesForPrefix(t *testing.T) {
	tests := []timeZonesTestCases{
		{
			num:               "+442073238299",
			expectedTimeZones: []string{"Europe/London"},
		},
		{
			num:               "+61491570156",
			expectedTimeZones: []string{"Australia/Sydney"},
		},
		{
			num:               "+61255501234",
			expectedTimeZones: []string{"Australia/Sydney"},
		},
		{
			num:               "+12067798181",
			expectedTimeZones: []string{"America/Los_Angeles"},
		},
		{
			num:               "+390399123456",
			expectedTimeZones: []string{"Europe/Rome"},
		},
		{
			num:               "+541151123456",
			expectedTimeZones: []string{"America/Buenos_Aires"},
		},
		{
			num:               "+15167706076",
			expectedTimeZones: []string{"America/New_York"},
		},
		{
			num:               "+917999999543",
			expectedTimeZones: []string{"Asia/Calcutta"},
		},
		{
			num:               "+540111561234567",
			expectedTimeZones: []string{"America/Buenos_Aires"},
		},
		{
			num:               "+18504320800",
			expectedTimeZones: []string{"America/Chicago"},
		},
		{
			num:               "+14079395277",
			expectedTimeZones: []string{"America/New_York"},
		},
		{
			num:               "+18508632167",
			expectedTimeZones: []string{"America/Chicago"},
		},
		{
			num:               "+40213158207",
			expectedTimeZones: []string{"Europe/Bucharest"},
		},
		// UTC +5:45
		{
			num:               "+97714240520",
			expectedTimeZones: []string{"Asia/Katmandu"},
		},
		// UTC -3:30
		{
			num:               "+17097264534",
			expectedTimeZones: []string{"America/St_Johns"},
		},
		{
			num:               "0000000000",
			expectedTimeZones: []string{"Etc/Unknown"},
		},
		{
			num:               "+31112",
			expectedTimeZones: []string{"Europe/Amsterdam"},
		},
		{
			num:               "+6837000",
			expectedTimeZones: []string{"Pacific/Niue"},
		},
		{
			num: "+1911",
			expectedTimeZones: []string{
				"America/Adak", "America/Anchorage", "America/Anguilla", "America/Antigua",
				"America/Barbados", "America/Boise", "America/Cayman", "America/Chicago",
				"America/Denver", "America/Dominica", "America/Edmonton", "America/Fort_Nelson",
				"America/Grand_Turk", "America/Grenada", "America/Halifax", "America/Jamaica",
				"America/Juneau", "America/Los_Angeles", "America/Lower_Princes", "America/Montserrat",
				"America/Nassau", "America/New_York", "America/North_Dakota/Center",
				"America/Phoenix", "America/Port_of_Spain", "America/Puerto_Rico",
				"America/Regina", "America/Santo_Domingo", "America/St_Johns", "America/St_Kitts",
				"America/St_Lucia", "America/St_Thomas", "America/St_Vincent", "America/Toronto",
				"America/Tortola", "America/Vancouver", "America/Winnipeg", "Atlantic/Bermuda",
				"Pacific/Guam", "Pacific/Honolulu", "Pacific/Pago_Pago", "Pacific/Saipan",
			},
		},
	}

	for _, test := range tests {
		timeZones, err := GetTimezonesForPrefix(test.num)
		if err != nil {
			t.Errorf("Failed to getTimezone for the number %s: %s", test.num, err)
		}

		if len(timeZones) == 0 {
			t.Errorf("Expected at least 1 timezone.")
		}

		if !reflect.DeepEqual(timeZones, test.expectedTimeZones) {
			t.Errorf("Expected '%v', got '%v' for '%s'", test.expectedTimeZones, timeZones, test.num)
		}

		num, err := Parse(test.num, "")
		if err != nil {
			continue
		}

		timeZones, err = GetTimezonesForNumber(num)
		if err != nil {
			t.Errorf("Failed to getTimezone for the number %s: %s", num, err)
		}

		if len(timeZones) == 0 {
			t.Errorf("Expected at least 1 timezone.")
		}

		if !reflect.DeepEqual(timeZones, test.expectedTimeZones) {
			t.Errorf("Expected '%v', got '%v' for '%s'", test.expectedTimeZones, timeZones, num)
		}
	}
}

func TestGetCarrierForNumber(t *testing.T) {
	tests := []struct {
		num      string
		lang     string
		expected string
	}{
		{num: "+8613702032331", lang: "en", expected: "China Mobile"},
		{num: "+8613702032331", lang: "zh", expected: "中国移动"},
		{num: "+6281377468527", lang: "en", expected: "Telkomsel"},
		{num: "+8613323241342", lang: "en", expected: "China Telecom"},
		{num: "+61491570156", lang: "en", expected: "Telstra"},
		{num: "+917999999543", lang: "en", expected: "Reliance Jio"},
		{num: "+593992218722", lang: "en", expected: "Claro"},
	}
	for _, test := range tests {
		number, err := Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		carrier, err := GetCarrierForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to getCarrier for the number %s: %s", test.num, err)
		}
		if test.expected != carrier {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expected, carrier, test.num)
		}
	}
}

func TestGetCarrierWithPrefixForNumber(t *testing.T) {
	tests := []struct {
		num             string
		lang            string
		expectedCarrier string
		expectedPrefix  int
	}{
		{num: "+8613702032331", lang: "en", expectedCarrier: "China Mobile", expectedPrefix: 86137},
		{num: "+8613702032331", lang: "zh", expectedCarrier: "中国移动", expectedPrefix: 86137},
		{num: "+6281377468527", lang: "en", expectedCarrier: "Telkomsel", expectedPrefix: 62813},
		{num: "+8613323241342", lang: "en", expectedCarrier: "China Telecom", expectedPrefix: 86133},
		{num: "+61491570156", lang: "en", expectedCarrier: "Telstra", expectedPrefix: 6149},
		{num: "+917999999543", lang: "en", expectedCarrier: "Reliance Jio", expectedPrefix: 917999},
		{num: "+593992218722", lang: "en", expectedCarrier: "Claro", expectedPrefix: 5939922},
		{num: "+201987654321", lang: "en", expectedCarrier: "", expectedPrefix: 0},
		{num: "+201987654321", lang: "notFound", expectedCarrier: "", expectedPrefix: 0},
	}
	for _, test := range tests {
		number, err := Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		carrier, prefix, err := GetCarrierWithPrefixForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to getCarrierWithPrefix for the number %s: %s", test.num, err)
		}
		if test.expectedCarrier != carrier {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expectedCarrier, carrier, test.num)
		}
		if test.expectedPrefix != prefix {
			t.Errorf("Expected '%d', got '%d' for '%s'", test.expectedPrefix, prefix, test.num)
		}
	}
}

func TestGetCarrierWithPrefixForNumberWithConcurrency(t *testing.T) {
	number, _ := Parse("+8613702032331", "ZZ")

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := GetCarrierWithPrefixForNumber(number, "en")
			if err != nil {
				t.Errorf("Failed to getCarrierWithPrefix for the number %s: %s", "+8613702032331", err)
			}
		}()
	}

	wg.Wait()
}

func TestGetGeocodingForNumber(t *testing.T) {
	tests := []struct {
		num      string
		lang     string
		expected string
	}{
		{num: "+8613702032331", lang: "en", expected: "Tianjin"},
		{num: "+8613702032331", lang: "zh", expected: "天津市"},
		{num: "+863197785050", lang: "zh", expected: "河北省邢台市"},
		{num: "+8613323241342", lang: "en", expected: "Baoding, Hebei"},
		{num: "+917999499543", lang: "en", expected: "Ahmedabad Local, Gujarat"},
		{num: "+17047181840", lang: "en", expected: "North Carolina"},
		{num: "+12542462158", lang: "en", expected: "Texas"},
		{num: "+16193165996", lang: "en", expected: "California"},
		{num: "+12067799191", lang: "en", expected: "Washington State"},
		{num: "+447825602614", lang: "en", expected: "United Kingdom"},
	}
	for _, test := range tests {
		number, err := Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		geocoding, err := GetGeocodingForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to getGeocoding for the number %s: %s", test.num, err)
		}
		if test.expected != geocoding {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expected, geocoding, test.num)
		}
	}
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

func TestMaybeSeparateExtensionFromPhone(t *testing.T) {
	tests := []struct {
		name      string
		rawPhone  string
		wantPhone string
		wantExt   string
	}{
		{
			name:      "Blank",
			rawPhone:  "",
			wantPhone: "",
			wantExt:   "",
		},
		{
			name:      "Number only",
			rawPhone:  "+1 306-555-1234",
			wantPhone: "+1 306-555-1234",
			wantExt:   "",
		},
		{
			name:      "Local phone",
			rawPhone:  "555-1234",
			wantPhone: "555-1234",
			wantExt:   "",
		},
		{
			name:      "Local phone with ext",
			rawPhone:  "555-1234 x 78",
			wantPhone: "555-1234",
			wantExt:   "78",
		},
		{
			name:      "Wait separator",
			rawPhone:  "+1 306-555-1234;78",
			wantPhone: "+1 306-555-1234",
			wantExt:   "78",
		},
		{
			name:      "Pause separator",
			rawPhone:  "+1 306-555-1234,78",
			wantPhone: "+1 306-555-1234",
			wantExt:   "78",
		},
		{
			name:      "Extra space separator",
			rawPhone:  "+1 306-555-1234   ,  78",
			wantPhone: "+1 306-555-1234",
			wantExt:   "78",
		},
		{
			name:      "ext. separator",
			rawPhone:  "+1 306-555-1234 ext. 78",
			wantPhone: "+1 306-555-1234",
			wantExt:   "78",
		},
		{
			name:      "# separator",
			rawPhone:  "+1 306-555-1234 #78",
			wantPhone: "+1 306-555-1234",
			wantExt:   "78",
		},
		{
			name:      "RFC3966",
			rawPhone:  "+1 306-555-1234;ext=78",
			wantPhone: "+1 306-555-1234",
			wantExt:   "78",
		},
		{
			name:      "Vanity phone number",
			rawPhone:  "+1 800-Go-Green #78",
			wantPhone: "+1 800-Go-Green",
			wantExt:   "78",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MaybeSeparateExtensionFromPhone(tt.rawPhone)
			if got != tt.wantPhone {
				t.Errorf("Phone got = %v, want %v", got, tt.wantPhone)
			}
			if got1 != tt.wantExt {
				t.Errorf("Extension got = %v, want %v", got1, tt.wantExt)
			}
		})
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

func TestRegexCacheWrite(t *testing.T) {
	pattern1 := "TestRegexCacheWrite"
	if _, found1 := readFromRegexCache(pattern1); found1 {
		t.Errorf("pattern |%v| is in the cache", pattern1)
	}
	regex1 := regexFor(pattern1)
	cachedRegex1, found1 := readFromRegexCache(pattern1)
	if !found1 {
		t.Errorf("pattern |%v| is not in the cache", pattern1)
	}
	if regex1 != cachedRegex1 {
		t.Error("expected the same instance, but got a different one")
	}
	pattern2 := pattern1 + "."
	if _, found2 := readFromRegexCache(pattern2); found2 {
		t.Errorf("pattern |%v| is in the cache", pattern2)
	}
}

func TestRegexCacheRead(t *testing.T) {
	pattern1 := "TestRegexCacheRead"
	if _, found1 := readFromRegexCache(pattern1); found1 {
		t.Errorf("pattern |%v| is in the cache", pattern1)
	}
	regex1 := regexp.MustCompile(pattern1)
	writeToRegexCache(pattern1, regex1)
	if cachedRegex1 := regexFor(pattern1); cachedRegex1 != regex1 {
		t.Error("expected the same instance, but got a different one")
	}
	cachedRegex1, found1 := readFromRegexCache(pattern1)
	if !found1 {
		t.Errorf("pattern |%v| is not in the cache", pattern1)
	}
	if cachedRegex1 != regex1 {
		t.Error("expected the same instance, but got a different one")
	}
	pattern2 := pattern1 + "."
	if _, found2 := readFromRegexCache(pattern2); found2 {
		t.Errorf("pattern |%v| is in the cache", pattern2)
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

func TestGetSafeCarrierDisplayNameForNumber(t *testing.T) {
	tests := []struct {
		num      string
		lang     string
		expected string
	}{
		{num: "+447387654321", lang: "en", expected: ""},
		{num: "+244917654321", lang: "en", expected: "Movicel"},
	}
	for _, test := range tests {
		number, err := Parse(test.num, "ZZ")
		if err != nil {
			t.Errorf("Failed to parse number %s: %s", test.num, err)
		}
		carrier, err := GetSafeCarrierDisplayNameForNumber(number, test.lang)
		if err != nil {
			t.Errorf("Failed to getSafeCarrierDisplayNameForNumber for the number %s: %s", test.num, err)
		}
		if test.expected != carrier {
			t.Errorf("Expected '%s', got '%s' for '%s'", test.expected, carrier, test.num)
		}
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

func BenchmarkLoadMetadata(b *testing.B) {
	for i := 0; i <= b.N; i++ {
		initMetadata()
	}
}

func BenchmarkGetCarrierWithPrefixForNumber(b *testing.B) {
	number, _ := Parse("+8613702032331", "ZZ")

	for n := 0; n < b.N; n++ {
		_, _, err := GetCarrierWithPrefixForNumber(number, "en")
		if err != nil {
			b.Errorf("Failed to getCarrierWithPrefix for the number %s: %s", "+8613702032331", err)
		}
	}
}
