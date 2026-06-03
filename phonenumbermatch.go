// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/PhoneNumberMatch.java from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
package phonenumbers

import "fmt"

// PhoneNumberMatch is the immutable match of a phone number within a piece of
// text. Matches may be found using FindNumbers.
//
// A match consists of the phone number (Number) as well as the start and end
// byte offsets of the corresponding substring of the searched text. Use
// RawString to obtain the matched substring. Note that, unlike upstream's Java
// (which uses UTF-16 char offsets), Start and End are byte offsets into the
// searched string, so text[match.Start():match.End()] == match.RawString().
type PhoneNumberMatch struct {
	start     int
	rawString string
	number    *PhoneNumber
}

// newPhoneNumberMatch creates a new match. start is the byte offset into the
// target text, rawString the matched substring, and number the parsed number.
func newPhoneNumberMatch(start int, rawString string, number *PhoneNumber) *PhoneNumberMatch {
	return &PhoneNumberMatch{start: start, rawString: rawString, number: number}
}

// Number returns the phone number matched by the receiver.
func (m *PhoneNumberMatch) Number() *PhoneNumber { return m.number }

// Start returns the start byte offset of the matched phone number within the searched text.
func (m *PhoneNumberMatch) Start() int { return m.start }

// End returns the exclusive end byte offset of the matched phone number within the searched text.
func (m *PhoneNumberMatch) End() int { return m.start + len(m.rawString) }

// RawString returns the raw substring matched as a phone number in the searched text.
func (m *PhoneNumberMatch) RawString() string { return m.rawString }

// String returns a human-readable representation of the match.
func (m *PhoneNumberMatch) String() string {
	return fmt.Sprintf("PhoneNumberMatch [%d,%d) %s", m.Start(), m.End(), m.rawString)
}
