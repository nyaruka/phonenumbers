package phonenumbers

// Test fixtures ported from PhoneNumberUtilTest.java's PhoneNumber constants,
// for the faithful upstream test port (run against the synthetic test metadata).
// As upstream notes, these return fresh values each call to avoid accidental
// mutation of shared state across tests. Last reconciled against: v9.0.32

import "google.golang.org/protobuf/proto"

func pn(countryCode int32, nationalNumber uint64) *PhoneNumber {
	return &PhoneNumber{CountryCode: proto.Int32(countryCode), NationalNumber: proto.Uint64(nationalNumber)}
}

func alphaNumericNumber() *PhoneNumber { return pn(1, 80074935247) }
func aeUAN() *PhoneNumber              { return pn(971, 600123456) }
func arMobile() *PhoneNumber           { return pn(54, 91187654321) }
func arNumber() *PhoneNumber           { return pn(54, 1187654321) }
func auNumber() *PhoneNumber           { return pn(61, 236618300) }
func bsMobile() *PhoneNumber           { return pn(1, 2423570000) }
func bsNumber() *PhoneNumber           { return pn(1, 2423651234) }
func coFixedLine() *PhoneNumber        { return pn(57, 6012345678) }

// deNumber is the same as the example number for DE in the metadata.
func deNumber() *PhoneNumber      { return pn(49, 30123456) }
func deShortNumber() *PhoneNumber { return pn(49, 1234) }
func gbMobile() *PhoneNumber      { return pn(44, 7912345678) }
func gbNumber() *PhoneNumber      { return pn(44, 2070313000) }
func itMobile() *PhoneNumber      { return pn(39, 345678901) }

func itNumber() *PhoneNumber {
	n := pn(39, 236618300)
	n.ItalianLeadingZero = proto.Bool(true)
	return n
}

func jpStarNumber() *PhoneNumber { return pn(81, 2345) }

// Numbers to test the formatting rules from Mexico.
func mxMobile1() *PhoneNumber { return pn(52, 12345678900) }
func mxMobile2() *PhoneNumber { return pn(52, 15512345678) }
func mxNumber1() *PhoneNumber { return pn(52, 3312345678) }
func mxNumber2() *PhoneNumber { return pn(52, 8211234567) }
func nzNumber() *PhoneNumber  { return pn(64, 33316005) }
func sgNumber() *PhoneNumber  { return pn(65, 65218000) }

// usLongNumber is a too-long and hence invalid US number.
func usLongNumber() *PhoneNumber { return pn(1, 65025300001) }
func usNumber() *PhoneNumber     { return pn(1, 6502530000) }
func usPremium() *PhoneNumber    { return pn(1, 9002530000) }

// Too short, but still possible US numbers.
func usLocalNumber() *PhoneNumber      { return pn(1, 2530000) }
func usShortByOneNumber() *PhoneNumber { return pn(1, 650253000) }
func usTollFree() *PhoneNumber         { return pn(1, 8002530000) }
func usSpoof() *PhoneNumber            { return pn(1, 0) }

func usSpoofWithRawInput() *PhoneNumber {
	n := pn(1, 0)
	n.RawInput = proto.String("000-000-0000")
	return n
}

func uzFixedLine() *PhoneNumber           { return pn(998, 612201234) }
func uzMobile() *PhoneNumber              { return pn(998, 950123456) }
func internationalTollFree() *PhoneNumber { return pn(800, 12345678) }

// internationalTollFreeTooLong is the same length as numbers for the other
// non-geographical country prefix in the test metadata, but is not valid
// because they differ in their country calling code.
func internationalTollFreeTooLong() *PhoneNumber { return pn(800, 123456789) }
func universalPremiumRate() *PhoneNumber         { return pn(979, 123456789) }
func unknownCountryCodeNoRawInput() *PhoneNumber { return pn(2, 12345) }
