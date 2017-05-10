package phonenumbers

// merge merges two number formats
func (nf *NumberFormat) merge(other *NumberFormat) {
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

func (pd *PhoneNumberDesc) hasPossibleLength(length int32) bool {
	for _, l := range pd.PossibleLength {
		if l == length {
			return true
		}
	}

	return false
}

func (pd *PhoneNumberDesc) hasPossibleLengthLocalOnly(length int32) bool {
	for _, l := range pd.PossibleLengthLocalOnly {
		if l == length {
			return true
		}
	}
	return false
}

// utility function called by the builder to remove possible lengths from our protobufs since
// we don't want to write them
func (pd *PhoneNumberDesc) clearPossibleLengths() {
	if pd == nil {
		return
	}
	pd.PossibleLength = nil
	pd.PossibleLengthLocalOnly = nil
}

// ClearPossibleLengths is called by the builder to remove possible lengths from our protobufs since
// we don't want to write them
func (md *PhoneMetadata) ClearPossibleLengths() {
	md.GeneralDesc.clearPossibleLengths()
	md.NoInternationalDialling.clearPossibleLengths()
	md.FixedLine.clearPossibleLengths()
	md.Mobile.clearPossibleLengths()
	md.Pager.clearPossibleLengths()
	md.TollFree.clearPossibleLengths()
	md.PremiumRate.clearPossibleLengths()
	md.SharedCost.clearPossibleLengths()
	md.PersonalNumber.clearPossibleLengths()
	md.Voip.clearPossibleLengths()
	md.Uan.clearPossibleLengths()
	md.Voicemail.clearPossibleLengths()
	md.StandardRate.clearPossibleLengths()
	md.ShortCode.clearPossibleLengths()
	md.Emergency.clearPossibleLengths()
	md.CarrierSpecific.clearPossibleLengths()
}
