package phonenumbers

// Faithful port of PhoneNumberUtilTest.java against synthetic test metadata,
// reconciled v9.0.32.
//
// Covers the number-type predicates (testIsPremiumRate, testIsTollFree, ...)
// and the number-component helpers (testGetNationalSignificantNumber,
// testGetLengthOfGeographicalAreaCode, ...). Each test activates the synthetic
// test metadata via useTestMetadata so expectations stay stable across
// real-world metadata refreshes.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestIsPremiumRate(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, PREMIUM_RATE, GetNumberType(usPremium()))

	premiumRateNumber := &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(892123)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	premiumRateNumber = &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(9187654321)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	premiumRateNumber = &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(9001654321)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	premiumRateNumber = &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(90091234567)}
	assert.Equal(t, PREMIUM_RATE, GetNumberType(premiumRateNumber))

	assert.Equal(t, PREMIUM_RATE, GetNumberType(universalPremiumRate()))
}

func TestIsTollFree(t *testing.T) {
	useTestMetadata(t)

	tollFreeNumber := &PhoneNumber{CountryCode: proto.Int32(1), NationalNumber: proto.Uint64(8881234567)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	tollFreeNumber = &PhoneNumber{CountryCode: proto.Int32(39), NationalNumber: proto.Uint64(803123)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	tollFreeNumber = &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(8012345678)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	tollFreeNumber = &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(8001234567)}
	assert.Equal(t, TOLL_FREE, GetNumberType(tollFreeNumber))

	assert.Equal(t, TOLL_FREE, GetNumberType(internationalTollFree()))
}

func TestIsMobile(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, MOBILE, GetNumberType(bsMobile()))
	assert.Equal(t, MOBILE, GetNumberType(gbMobile()))
	assert.Equal(t, MOBILE, GetNumberType(itMobile()))
	assert.Equal(t, MOBILE, GetNumberType(arMobile()))

	mobileNumber := &PhoneNumber{CountryCode: proto.Int32(49), NationalNumber: proto.Uint64(15123456789)}
	assert.Equal(t, MOBILE, GetNumberType(mobileNumber))
}

func TestIsFixedLine(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, FIXED_LINE, GetNumberType(bsNumber()))
	assert.Equal(t, FIXED_LINE, GetNumberType(itNumber()))
	assert.Equal(t, FIXED_LINE, GetNumberType(gbNumber()))
	assert.Equal(t, FIXED_LINE, GetNumberType(deNumber()))
}

func TestIsFixedLineAndMobile(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, FIXED_LINE_OR_MOBILE, GetNumberType(usNumber()))

	fixedLineAndMobileNumber := &PhoneNumber{CountryCode: proto.Int32(54), NationalNumber: proto.Uint64(1987654321)}
	assert.Equal(t, FIXED_LINE_OR_MOBILE, GetNumberType(fixedLineAndMobileNumber))
}

func TestIsSharedCost(t *testing.T) {
	useTestMetadata(t)
	gbNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(8431231234)}
	assert.Equal(t, SHARED_COST, GetNumberType(gbNumber))
}

func TestIsVoip(t *testing.T) {
	useTestMetadata(t)
	gbNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(5631231234)}
	assert.Equal(t, VOIP, GetNumberType(gbNumber))
}

func TestIsPersonalNumber(t *testing.T) {
	useTestMetadata(t)
	gbNumber := &PhoneNumber{CountryCode: proto.Int32(44), NationalNumber: proto.Uint64(7031231234)}
	assert.Equal(t, PERSONAL_NUMBER, GetNumberType(gbNumber))
}

func TestIsUnknown(t *testing.T) {
	useTestMetadata(t)
	// Invalid numbers should be of type UNKNOWN.
	assert.Equal(t, UNKNOWN, GetNumberType(usLocalNumber()))
}

func TestGetNationalSignificantNumber(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, "6502530000", GetNationalSignificantNumber(usNumber()))

	// An Italian mobile number.
	assert.Equal(t, "345678901", GetNationalSignificantNumber(itMobile()))

	// An Italian fixed line number.
	assert.Equal(t, "0236618300", GetNationalSignificantNumber(itNumber()))

	assert.Equal(t, "12345678", GetNationalSignificantNumber(internationalTollFree()))
}

func TestGetNationalSignificantNumberManyLeadingZeros(t *testing.T) {
	useTestMetadata(t)
	number := &PhoneNumber{
		CountryCode:          proto.Int32(1),
		NationalNumber:       proto.Uint64(650),
		ItalianLeadingZero:   proto.Bool(true),
		NumberOfLeadingZeros: proto.Int32(2),
	}
	assert.Equal(t, "00650", GetNationalSignificantNumber(number))

	// Set a bad value; we shouldn't crash, we shouldn't output any leading zeros at all.
	number.NumberOfLeadingZeros = proto.Int32(-3)
	assert.Equal(t, "650", GetNationalSignificantNumber(number))
}

func TestGetLengthOfGeographicalAreaCode(t *testing.T) {
	useTestMetadata(t)
	// Google MTV, which has area code "650".
	assert.Equal(t, 3, GetLengthOfGeographicalAreaCode(usNumber()))

	// A North America toll-free number, which has no area code.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(usTollFree()))

	// Google London, which has area code "20".
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(gbNumber()))

	// A mobile number in the UK does not have an area code (by default, mobile numbers do not,
	// unless they have been added to our list of exceptions).
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(gbMobile()))

	// Google Buenos Aires, which has area code "11".
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(arNumber()))

	// A mobile number in Argentina also has an area code.
	assert.Equal(t, 3, GetLengthOfGeographicalAreaCode(arMobile()))

	// Google Sydney, which has area code "2".
	assert.Equal(t, 1, GetLengthOfGeographicalAreaCode(auNumber()))

	// Italian numbers - there is no national prefix, but it still has an area code.
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(itNumber()))

	// Mexico numbers - there is no national prefix, but it still has an area code.
	assert.Equal(t, 2, GetLengthOfGeographicalAreaCode(mxNumber1()))

	// Google Singapore. Singapore has no area code and no national prefix.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(sgNumber()))

	// An invalid US number (1 digit shorter), which has no area code.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(usShortByOneNumber()))

	// An international toll free number, which has no area code.
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(internationalTollFree()))

	// A mobile number from China is geographical, but does not have an area code.
	cnMobile := &PhoneNumber{CountryCode: proto.Int32(86), NationalNumber: proto.Uint64(18912341234)}
	assert.Equal(t, 0, GetLengthOfGeographicalAreaCode(cnMobile))
}

func TestGetLengthOfNationalDestinationCode(t *testing.T) {
	useTestMetadata(t)
	// Google MTV, which has national destination code (NDC) "650".
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(usNumber()))

	// A North America toll-free number, which has NDC "800".
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(usTollFree()))

	// Google London, which has NDC "20".
	assert.Equal(t, 2, GetLengthOfNationalDestinationCode(gbNumber()))

	// A UK mobile phone, which has NDC "7912".
	assert.Equal(t, 4, GetLengthOfNationalDestinationCode(gbMobile()))

	// Google Buenos Aires, which has NDC "11".
	assert.Equal(t, 2, GetLengthOfNationalDestinationCode(arNumber()))

	// An Argentinian mobile which has NDC "911".
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(arMobile()))

	// Google Sydney, which has NDC "2".
	assert.Equal(t, 1, GetLengthOfNationalDestinationCode(auNumber()))

	// Google Singapore, which has NDC "6521".
	assert.Equal(t, 4, GetLengthOfNationalDestinationCode(sgNumber()))

	// An invalid US number (1 digit shorter), which has no NDC.
	assert.Equal(t, 0, GetLengthOfNationalDestinationCode(usShortByOneNumber()))

	// A number containing an invalid country calling code, which shouldn't have any NDC.
	number := &PhoneNumber{CountryCode: proto.Int32(123), NationalNumber: proto.Uint64(6502530000)}
	assert.Equal(t, 0, GetLengthOfNationalDestinationCode(number))

	// An international toll free number, which has NDC "1234".
	assert.Equal(t, 4, GetLengthOfNationalDestinationCode(internationalTollFree()))

	// A mobile number from China is geographical, but does not have an area code: however it still
	// can be considered to have a national destination code.
	cnMobile := &PhoneNumber{CountryCode: proto.Int32(86), NationalNumber: proto.Uint64(18912341234)}
	assert.Equal(t, 3, GetLengthOfNationalDestinationCode(cnMobile))
}

func TestGetCountryMobileToken(t *testing.T) {
	useTestMetadata(t)
	assert.Equal(t, "9", GetCountryMobileToken(GetCountryCodeForRegion(regionCode.AR)))

	// Country calling code for Sweden, which has no mobile token.
	assert.Equal(t, "", GetCountryMobileToken(GetCountryCodeForRegion(regionCode.SE)))
}

func TestIsNumberGeographical(t *testing.T) {
	useTestMetadata(t)
	assert.False(t, isNumberGeographical(bsMobile()))              // Bahamas, mobile phone number.
	assert.True(t, isNumberGeographical(auNumber()))               // Australian fixed line number.
	assert.False(t, isNumberGeographical(internationalTollFree())) // International toll free number.
	// We test that mobile phone numbers in relevant regions are indeed considered geographical.
	assert.True(t, isNumberGeographical(arMobile()))  // Argentina, mobile phone number.
	assert.True(t, isNumberGeographical(mxMobile1())) // Mexico, mobile phone number.
	assert.True(t, isNumberGeographical(mxMobile2())) // Mexico, another mobile phone number.
}
