// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/PhoneNumberUtil.java (enums).
package phonenumbers

// INTERNATIONAL and NATIONAL formats are consistent with the definition
// in ITU-T Recommendation E123. For example, the number of the Google
// Switzerland office will be written as "+41 44 668 1800" in
// INTERNATIONAL format, and as "044 668 1800" in NATIONAL format. E164
// format is as per INTERNATIONAL format but with no formatting applied,
// e.g. "+41446681800". RFC3966 is as per INTERNATIONAL format, but with
// all spaces and other separating symbols replaced with a hyphen, and
// with any phone number extension appended with ";ext=". It also will
// have a prefix of "tel:" added, e.g. "tel:+41-44-668-1800".
//
// Note: If you are considering storing the number in a neutral format,
// you are highly advised to use the PhoneNumber class.

type PhoneNumberFormat int

const (
	E164 PhoneNumberFormat = iota
	INTERNATIONAL
	NATIONAL
	RFC3966
)

type PhoneNumberType int

const (
	// NOTES:
	//
	// FIXED_LINE_OR_MOBILE:
	//     In some regions (e.g. the USA), it is impossible to distinguish
	//     between fixed-line and mobile numbers by looking at the phone
	//     number itself.
	// SHARED_COST:
	//     The cost of this call is shared between the caller and the
	//     recipient, and is hence typically less than PREMIUM_RATE calls.
	//     See // http://en.wikipedia.org/wiki/Shared_Cost_Service for
	//     more information.
	// VOIP:
	//     Voice over IP numbers. This includes TSoIP (Telephony Service over IP).
	// PERSONAL_NUMBER:
	//     A personal number is associated with a particular person, and may
	//     be routed to either a MOBILE or FIXED_LINE number. Some more
	//     information can be found here:
	//     http://en.wikipedia.org/wiki/Personal_Numbers
	// UAN:
	//     Used for "Universal Access Numbers" or "Company Numbers". They
	//     may be further routed to specific offices, but allow one number
	//     to be used for a company.
	// VOICEMAIL:
	//     Used for "Voice Mail Access Numbers".
	// UNKNOWN:
	//     A phone number is of type UNKNOWN when it does not fit any of
	// the known patterns for a specific region.
	FIXED_LINE PhoneNumberType = iota
	MOBILE
	FIXED_LINE_OR_MOBILE
	TOLL_FREE
	PREMIUM_RATE
	SHARED_COST
	VOIP
	PERSONAL_NUMBER
	PAGER
	UAN
	VOICEMAIL
	UNKNOWN
)

type MatchType int

const (
	NOT_A_NUMBER MatchType = iota
	NO_MATCH
	SHORT_NSN_MATCH
	NSN_MATCH
	EXACT_MATCH
)

type ValidationResult int

const (
	IS_POSSIBLE ValidationResult = iota
	IS_POSSIBLE_LOCAL_ONLY
	INVALID_COUNTRY_CODE
	TOO_SHORT
	INVALID_LENGTH
	TOO_LONG
)

// TODO(ttacon): leniency comments?
type Leniency int

const (
	POSSIBLE Leniency = iota
	VALID
	STRICT_GROUPING
	EXACT_GROUPING
)

func (l Leniency) Verify(number *PhoneNumber, candidate string) bool {

	switch l {
	case POSSIBLE:
		return IsPossibleNumber(number)
	case VALID:
		if !IsValidNumber(number) ||
			!containsOnlyValidXChars(number, candidate) {
			return false
		}
		return isNationalPrefixPresentIfRequired(number)
	case STRICT_GROUPING:
		if !IsValidNumber(number) ||
			!containsOnlyValidXChars(number, candidate) ||
			containsMoreThanOneSlashInNationalNumber(number, candidate) ||
			!isNationalPrefixPresentIfRequired(number) {
			return false
		}
		return checkNumberGroupingIsValid(number, candidate,
			func(number *PhoneNumber,
				normalizedCandidate string,
				expectedNumberGroups []string) bool {
				return allNumberGroupsRemainGrouped(
					number, normalizedCandidate, expectedNumberGroups)
			})
	case EXACT_GROUPING:
		if !IsValidNumber(number) ||
			!containsOnlyValidXChars(number, candidate) ||
			containsMoreThanOneSlashInNationalNumber(number, candidate) ||
			!isNationalPrefixPresentIfRequired(number) {
			return false
		}
		return checkNumberGroupingIsValid(number, candidate,
			func(number *PhoneNumber,
				normalizedCandidate string,
				expectedNumberGroups []string) bool {
				return allNumberGroupsAreExactlyPresent(
					number, normalizedCandidate, expectedNumberGroups)
			})
	}
	return false
}
