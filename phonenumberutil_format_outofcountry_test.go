package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java
// out-of-country / preferred-carrier formatting tests, run against the synthetic
// test metadata. Method names and assertions mirror the Java. Last reconciled
// against: v9.0.32

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestFormatOutOfCountryCallingNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "00 1 900 253 0000", FormatOutOfCountryCallingNumber(usPremium(), regionCode.DE))
	assert.Equal(t, "1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.BS))
	assert.Equal(t, "00 1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.PL))
	assert.Equal(t, "011 44 7912 345 678", FormatOutOfCountryCallingNumber(gbMobile(), regionCode.US))
	assert.Equal(t, "00 49 1234", FormatOutOfCountryCallingNumber(deShortNumber(), regionCode.GB))
	// Note this number is correctly formatted without national prefix.
	assert.Equal(t, "1234", FormatOutOfCountryCallingNumber(deShortNumber(), regionCode.DE))

	assert.Equal(t, "011 39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.US))
	assert.Equal(t, "02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.IT))
	assert.Equal(t, "+39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.SG))

	assert.Equal(t, "6521 8000", FormatOutOfCountryCallingNumber(sgNumber(), regionCode.SG))

	assert.Equal(t, "011 54 9 11 8765 4321", FormatOutOfCountryCallingNumber(arMobile(), regionCode.US))
	assert.Equal(t, "011 800 1234 5678", FormatOutOfCountryCallingNumber(internationalTollFree(), regionCode.US))

	arNumberWithExtn := arMobile()
	arNumberWithExtn.Extension = proto.String("1234")
	assert.Equal(t, "011 54 9 11 8765 4321 ext. 1234", FormatOutOfCountryCallingNumber(arNumberWithExtn, regionCode.US))
	assert.Equal(t, "0011 54 9 11 8765 4321 ext. 1234", FormatOutOfCountryCallingNumber(arNumberWithExtn, regionCode.AU))
	assert.Equal(t, "011 15 8765-4321 ext. 1234", FormatOutOfCountryCallingNumber(arNumberWithExtn, regionCode.AR))
}

func TestFormatOutOfCountryWithInvalidRegion(t *testing.T) {
	useTestMetadata(t)

	// AQ/Antarctica isn't a valid region code for phone number formatting, so this falls back to
	// intl formatting.
	assert.Equal(t, "+1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), "AQ"))
	// For region code 001, the out-of-country format always turns into the international format.
	assert.Equal(t, "+1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.UN001))
}

func TestFormatOutOfCountryWithPreferredIntlPrefix(t *testing.T) {
	useTestMetadata(t)

	// This should use 0011, since that is the preferred international prefix (both 0011 and 0012
	// are accepted as possible international prefixes in our test metadata.)
	assert.Equal(t, "0011 39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.AU))
	// Testing preferred international prefixes with ~ are supported (designates waiting).
	assert.Equal(t, "8~10 39 02 3661 8300", FormatOutOfCountryCallingNumber(itNumber(), regionCode.UZ))
}

func TestFormatWithPreferredCarrierCode(t *testing.T) {
	useTestMetadata(t)

	// We only support this for AR in our test metadata.
	arNumber := pn(54, 91234125678)
	// Test formatting with no preferred carrier code stored in the number itself.
	assert.Equal(t, "01234 15 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	assert.Equal(t, "01234 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, ""))
	// Test formatting with preferred carrier code present.
	arNumber.PreferredDomesticCarrierCode = proto.String("19")
	assert.Equal(t, "01234 12-5678", Format(arNumber, NATIONAL))
	assert.Equal(t, "01234 19 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	assert.Equal(t, "01234 19 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, ""))
	// When the preferred_domestic_carrier_code is present (even when it is just a space), use it
	// instead of the default carrier code passed in.
	arNumber.PreferredDomesticCarrierCode = proto.String(" ")
	assert.Equal(t, "01234   12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	// When the preferred_domestic_carrier_code is present but empty, treat it as unset and use
	// instead the default carrier code passed in.
	arNumber.PreferredDomesticCarrierCode = proto.String("")
	assert.Equal(t, "01234 15 12-5678", FormatNationalNumberWithPreferredCarrierCode(arNumber, "15"))
	// We don't support this for the US so there should be no change.
	usNumber := pn(1, 4241231234)
	usNumber.PreferredDomesticCarrierCode = proto.String("99")
	assert.Equal(t, "424 123 1234", Format(usNumber, NATIONAL))
	assert.Equal(t, "424 123 1234", FormatNationalNumberWithPreferredCarrierCode(usNumber, "15"))
}
