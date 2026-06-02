package phonenumbers

// Faithful port of upstream libphonenumber's PhoneNumberUtilTest.java formatting
// tests, run against the synthetic test metadata (see testmetadata_test.go).
// Method names and assertions mirror the Java so this file can be kept in sync
// with upstream.
//
// Note: testFormatWithCarrierCode is already covered by
// TestFormatWithCarrierCodeTestMetadata in phonenumberutil_format_natprefix_test.go;
// the out-of-country / preferred-carrier / mobile-dialing / in-original-format
// methods are ported separately.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestFormatUSNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "650 253 0000", Format(usNumber(), NATIONAL))
	assert.Equal(t, "+1 650 253 0000", Format(usNumber(), INTERNATIONAL))

	assert.Equal(t, "800 253 0000", Format(usTollFree(), NATIONAL))
	assert.Equal(t, "+1 800 253 0000", Format(usTollFree(), INTERNATIONAL))

	assert.Equal(t, "900 253 0000", Format(usPremium(), NATIONAL))
	assert.Equal(t, "+1 900 253 0000", Format(usPremium(), INTERNATIONAL))
	assert.Equal(t, "tel:+1-900-253-0000", Format(usPremium(), RFC3966))
	// Numbers with all zeros in the national number part will be formatted by using the raw_input
	// if that is available no matter which format is specified.
	assert.Equal(t, "000-000-0000", Format(usSpoofWithRawInput(), NATIONAL))
	assert.Equal(t, "0", Format(usSpoof(), NATIONAL))
}

func TestFormatBSNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "242 365 1234", Format(bsNumber(), NATIONAL))
	assert.Equal(t, "+1 242 365 1234", Format(bsNumber(), INTERNATIONAL))
}

func TestFormatGBNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "(020) 7031 3000", Format(gbNumber(), NATIONAL))
	assert.Equal(t, "+44 20 7031 3000", Format(gbNumber(), INTERNATIONAL))

	assert.Equal(t, "(07912) 345 678", Format(gbMobile(), NATIONAL))
	assert.Equal(t, "+44 7912 345 678", Format(gbMobile(), INTERNATIONAL))
}

func TestFormatDENumber(t *testing.T) {
	useTestMetadata(t)

	deNumber := pn(49, 301234)
	assert.Equal(t, "030/1234", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 30/1234", Format(deNumber, INTERNATIONAL))
	assert.Equal(t, "tel:+49-30-1234", Format(deNumber, RFC3966))

	deNumber = pn(49, 291123)
	assert.Equal(t, "0291 123", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 291 123", Format(deNumber, INTERNATIONAL))

	deNumber = pn(49, 29112345678)
	assert.Equal(t, "0291 12345678", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 291 12345678", Format(deNumber, INTERNATIONAL))

	deNumber = pn(49, 912312345)
	assert.Equal(t, "09123 12345", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 9123 12345", Format(deNumber, INTERNATIONAL))

	deNumber = pn(49, 80212345)
	assert.Equal(t, "08021 2345", Format(deNumber, NATIONAL))
	assert.Equal(t, "+49 8021 2345", Format(deNumber, INTERNATIONAL))
	// Note this number is correctly formatted without national prefix. Most of the numbers that
	// are treated as invalid numbers by the library are short numbers, and they are usually not
	// dialed with national prefix.
	assert.Equal(t, "1234", Format(deShortNumber(), NATIONAL))
	assert.Equal(t, "+49 1234", Format(deShortNumber(), INTERNATIONAL))

	deNumber = pn(49, 41341234)
	assert.Equal(t, "04134 1234", Format(deNumber, NATIONAL))
}

func TestFormatITNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "02 3661 8300", Format(itNumber(), NATIONAL))
	assert.Equal(t, "+39 02 3661 8300", Format(itNumber(), INTERNATIONAL))
	assert.Equal(t, "+390236618300", Format(itNumber(), E164))

	assert.Equal(t, "345 678 901", Format(itMobile(), NATIONAL))
	assert.Equal(t, "+39 345 678 901", Format(itMobile(), INTERNATIONAL))
	assert.Equal(t, "+39345678901", Format(itMobile(), E164))
}

func TestFormatAUNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "02 3661 8300", Format(auNumber(), NATIONAL))
	assert.Equal(t, "+61 2 3661 8300", Format(auNumber(), INTERNATIONAL))
	assert.Equal(t, "+61236618300", Format(auNumber(), E164))

	auNumber := pn(61, 1800123456)
	assert.Equal(t, "1800 123 456", Format(auNumber, NATIONAL))
	assert.Equal(t, "+61 1800 123 456", Format(auNumber, INTERNATIONAL))
	assert.Equal(t, "+611800123456", Format(auNumber, E164))
}

func TestFormatARNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "011 8765-4321", Format(arNumber(), NATIONAL))
	assert.Equal(t, "+54 11 8765-4321", Format(arNumber(), INTERNATIONAL))
	assert.Equal(t, "+541187654321", Format(arNumber(), E164))

	assert.Equal(t, "011 15 8765-4321", Format(arMobile(), NATIONAL))
	assert.Equal(t, "+54 9 11 8765 4321", Format(arMobile(), INTERNATIONAL))
	assert.Equal(t, "+5491187654321", Format(arMobile(), E164))
}

func TestFormatMXNumber(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "045 234 567 8900", Format(mxMobile1(), NATIONAL))
	assert.Equal(t, "+52 1 234 567 8900", Format(mxMobile1(), INTERNATIONAL))
	assert.Equal(t, "+5212345678900", Format(mxMobile1(), E164))

	assert.Equal(t, "045 55 1234 5678", Format(mxMobile2(), NATIONAL))
	assert.Equal(t, "+52 1 55 1234 5678", Format(mxMobile2(), INTERNATIONAL))
	assert.Equal(t, "+5215512345678", Format(mxMobile2(), E164))

	assert.Equal(t, "01 33 1234 5678", Format(mxNumber1(), NATIONAL))
	assert.Equal(t, "+52 33 1234 5678", Format(mxNumber1(), INTERNATIONAL))
	assert.Equal(t, "+523312345678", Format(mxNumber1(), E164))

	assert.Equal(t, "01 821 123 4567", Format(mxNumber2(), NATIONAL))
	assert.Equal(t, "+52 821 123 4567", Format(mxNumber2(), INTERNATIONAL))
	assert.Equal(t, "+528211234567", Format(mxNumber2(), E164))
}

func TestFormatByPattern(t *testing.T) {
	useTestMetadata(t)

	newNumFormat := &NumberFormat{
		Pattern: proto.String("(\\d{3})(\\d{3})(\\d{4})"),
		Format:  proto.String("($1) $2-$3"),
	}
	newNumberFormats := []*NumberFormat{newNumFormat}

	assert.Equal(t, "(650) 253-0000", FormatByPattern(usNumber(), NATIONAL, newNumberFormats))
	assert.Equal(t, "+1 (650) 253-0000", FormatByPattern(usNumber(), INTERNATIONAL, newNumberFormats))
	usNumber2 := pn(1, 6507129823)
	assert.Equal(t, "tel:+1-650-712-9823", FormatByPattern(usNumber2, RFC3966, newNumberFormats))

	// $NP is set to '1' for the US. Here we check that for other NANPA countries the US rules are
	// followed.
	newNumFormat.NationalPrefixFormattingRule = proto.String("$NP ($FG)")
	newNumFormat.Format = proto.String("$1 $2-$3")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "1 (242) 365-1234", FormatByPattern(bsNumber(), NATIONAL, newNumberFormats))
	assert.Equal(t, "+1 242 365-1234", FormatByPattern(bsNumber(), INTERNATIONAL, newNumberFormats))

	newNumFormat.Pattern = proto.String("(\\d{2})(\\d{5})(\\d{3})")
	newNumFormat.Format = proto.String("$1-$2 $3")
	newNumberFormats[0] = newNumFormat

	assert.Equal(t, "02-36618 300", FormatByPattern(itNumber(), NATIONAL, newNumberFormats))
	assert.Equal(t, "+39 02-36618 300", FormatByPattern(itNumber(), INTERNATIONAL, newNumberFormats))

	newNumFormat.NationalPrefixFormattingRule = proto.String("$NP$FG")
	newNumFormat.Pattern = proto.String("(\\d{2})(\\d{4})(\\d{4})")
	newNumFormat.Format = proto.String("$1 $2 $3")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "020 7031 3000", FormatByPattern(gbNumber(), NATIONAL, newNumberFormats))

	newNumFormat.NationalPrefixFormattingRule = proto.String("($NP$FG)")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "(020) 7031 3000", FormatByPattern(gbNumber(), NATIONAL, newNumberFormats))

	newNumFormat.NationalPrefixFormattingRule = proto.String("")
	newNumberFormats[0] = newNumFormat
	assert.Equal(t, "20 7031 3000", FormatByPattern(gbNumber(), NATIONAL, newNumberFormats))

	assert.Equal(t, "+44 20 7031 3000", FormatByPattern(gbNumber(), INTERNATIONAL, newNumberFormats))
}

func TestFormatE164Number(t *testing.T) {
	useTestMetadata(t)

	assert.Equal(t, "+16502530000", Format(usNumber(), E164))
	assert.Equal(t, "+4930123456", Format(deNumber(), E164))
	assert.Equal(t, "+80012345678", Format(internationalTollFree(), E164))
}

func TestFormatNumberWithExtension(t *testing.T) {
	useTestMetadata(t)

	nzNumber := nzNumber()
	nzNumber.Extension = proto.String("1234")
	// Uses default extension prefix:
	assert.Equal(t, "03-331 6005 ext. 1234", Format(nzNumber, NATIONAL))
	// Uses RFC 3966 syntax.
	assert.Equal(t, "tel:+64-3-331-6005;ext=1234", Format(nzNumber, RFC3966))
	// Extension prefix overridden in the territory information for the US:
	usNumberWithExtension := usNumber()
	usNumberWithExtension.Extension = proto.String("4567")
	assert.Equal(t, "650 253 0000 extn. 4567", Format(usNumberWithExtension, NATIONAL))
}

func TestCountryWithNoNumberDesc(t *testing.T) {
	useTestMetadata(t)

	// Andorra is a country where we don't have PhoneNumberDesc info in the metadata.
	adNumber := pn(376, 12345)
	assert.Equal(t, "+376 12345", Format(adNumber, INTERNATIONAL))
	assert.Equal(t, "+37612345", Format(adNumber, E164))
	assert.Equal(t, "12345", Format(adNumber, NATIONAL))
	assert.Equal(t, UNKNOWN, GetNumberType(adNumber))
	assert.False(t, IsValidNumber(adNumber))

	// Test dialing a US number from within Andorra.
	assert.Equal(t, "00 1 650 253 0000", FormatOutOfCountryCallingNumber(usNumber(), regionCode.AD))
}

func TestUnknownCountryCallingCode(t *testing.T) {
	useTestMetadata(t)

	assert.False(t, IsValidNumber(unknownCountryCodeNoRawInput()))
	// It's not very well defined as to what the E164 representation for a number with an invalid
	// country calling code is, but just prefixing the country code and national number is about
	// the best we can do.
	assert.Equal(t, "+212345", Format(unknownCountryCodeNoRawInput(), E164))
}
