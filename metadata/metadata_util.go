// Package metadata holds the phone-number metadata value types (a port of
// upstream's Phonemetadata) together with the hand-written helpers on them. It
// is the acyclic home for these types now that the metadata loader depends on
// them; the root phonenumbers package re-exports them as aliases.
package metadata

// Merge folds the set fields of other into nf: for each field other sets, its
// value overwrites nf's, and leading-digit patterns are appended.
func (nf *NumberFormat) Merge(other *NumberFormat) {
	if other.Pattern != nil {
		nf.Pattern = other.Pattern
	}
	if other.Format != nil {
		nf.Format = other.Format
	}
	for i := 0; i < len(other.LeadingDigitsPattern); i++ {
		nf.LeadingDigitsPattern = append(nf.LeadingDigitsPattern, other.LeadingDigitsPattern[i])
	}
	if other.NationalPrefixFormattingRule != nil {
		nf.NationalPrefixFormattingRule = other.NationalPrefixFormattingRule
	}
	if other.DomesticCarrierCodeFormattingRule != nil {
		nf.DomesticCarrierCodeFormattingRule = other.DomesticCarrierCodeFormattingRule
	}
	if other.NationalPrefixOptionalWhenFormatting != nil {
		nf.NationalPrefixOptionalWhenFormatting = other.NationalPrefixOptionalWhenFormatting
	}
}

// HasPossibleLength reports whether length is one of pd's possible lengths.
func (pd *PhoneNumberDesc) HasPossibleLength(length int32) bool {
	for _, l := range pd.PossibleLength {
		if l == length {
			return true
		}
	}

	return false
}

// HasPossibleLengthLocalOnly reports whether length is one of pd's local-only
// possible lengths.
func (pd *PhoneNumberDesc) HasPossibleLengthLocalOnly(length int32) bool {
	for _, l := range pd.PossibleLengthLocalOnly {
		if l == length {
			return true
		}
	}
	return false
}
