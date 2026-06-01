package phonenumbers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// Regression tests for the national-prefix / $FG formatting-rule fix in
// phonenumberutil.go. These cover the three places where a formatting rule's
// first-group token must be applied the way upstream libphonenumber does
// (Java's String.replace / Matcher.replaceFirst semantics):
//
//  1. FormatByPattern compiling a user-supplied $NP$FG rule,
//  2. formatNsnUsingPattern applying a metadata national-prefix rule, and
//  3. the carrier-code branch of formatNsnUsingPatternWithCarrier.
//
// The test-metadata cases run against the synthetic metadata so they never go
// stale; TestFormatByPatternFGProdMetadata reproduces the original bug against
// the real embedded metadata.

func TestFormatNationalPrefixFormattingRule(t *testing.T) {
	useTestMetadata(t)

	// AR/MX mobile formats whose first $-token is "$2", with national-prefix
	// rules "0$1"/"$1": the "$1" must expand to the matched token rather than
	// being substituted literally (which would emit format group 1's digit).
	assert.Equal(t, "011 15 8765-4321", Format(arMobile(), NATIONAL))
	assert.Equal(t, "045 234 567 8900", Format(mxMobile1(), NATIONAL))
	assert.Equal(t, "045 55 1234 5678", Format(mxMobile2(), NATIONAL))

	// User-supplied rules containing $FG, compiled by FormatByPattern. $NP is
	// "1" for NANPA regions and "0" for GB.
	bs := &NumberFormat{
		Pattern:                      proto.String(`(\d{3})(\d{3})(\d{4})`),
		Format:                       proto.String("$1 $2-$3"),
		NationalPrefixFormattingRule: proto.String("$NP ($FG)"),
	}
	assert.Equal(t, "1 (242) 365-1234", FormatByPattern(bsNumber(), NATIONAL, []*NumberFormat{bs}))

	gbRule := func(rule string) []*NumberFormat {
		return []*NumberFormat{{
			Pattern:                      proto.String(`(\d{2})(\d{4})(\d{4})`),
			Format:                       proto.String("$1 $2 $3"),
			NationalPrefixFormattingRule: proto.String(rule),
		}}
	}
	assert.Equal(t, "020 7031 3000", FormatByPattern(gbNumber(), NATIONAL, gbRule("$NP$FG")))
	assert.Equal(t, "(020) 7031 3000", FormatByPattern(gbNumber(), NATIONAL, gbRule("($NP$FG)")))
}

// TestFormatByPatternFGProdMetadata is the original bug repro against the real
// embedded metadata: a $NP$FG rule used to emit a literal backslash
// ("0\ 7031 3000") because Go's regexp replacer treated "$1" as a group
// reference into the groupless $FG pattern.
func TestFormatByPatternFGProdMetadata(t *testing.T) {
	num, err := Parse("+442070313000", "GB")
	assert.NoError(t, err)

	formats := []*NumberFormat{{
		Pattern:                      proto.String(`(\d{2})(\d{4})(\d{4})`),
		Format:                       proto.String("$1 $2 $3"),
		NationalPrefixFormattingRule: proto.String("$NP$FG"),
	}}
	assert.Equal(t, "020 7031 3000", FormatByPattern(num, NATIONAL, formats))
}

// TestFormatWithCarrierCodeTestMetadata mirrors upstream PhoneNumberUtilTest.
// testFormatWithCarrierCode against the synthetic test metadata. AR mobile
// numbers exercise a carrier format whose first $-token is "$2" and whose
// domesticCarrierCodeFormattingRule is "$NP$FG $CC" -> "0$1 $CC"; the "$1" must
// expand to the matched first token ("$2"), not be left literal (which would
// resolve to format group 1 and emit the wrong digit). This guards the
// carrier-code branch of formatNsnUsingPatternWithCarrier.
func TestFormatWithCarrierCodeTestMetadata(t *testing.T) {
	useTestMetadata(t)

	arMobile := newPhoneNumber(54, 92234654321)
	assert.Equal(t, "02234 65-4321", Format(arMobile, NATIONAL))
	// Here we force 14 as the carrier code.
	assert.Equal(t, "02234 14 65-4321", FormatNationalNumberWithCarrierCode(arMobile, "14"))
	// Here we force the number to be shown with no carrier code.
	assert.Equal(t, "02234 65-4321", FormatNationalNumberWithCarrierCode(arMobile, ""))
	// E164 ignores national/carrier formatting, so no carrier code is present.
	assert.Equal(t, "+5492234654321", Format(arMobile, E164))
}
