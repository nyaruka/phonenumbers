// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/ShortNumberInfo.java.
package phonenumbers

import (
	"slices"

	"github.com/nyaruka/phonenumbers/v2/internal/regexbasedmatcher"
	"github.com/nyaruka/phonenumbers/v2/internal/serialize"
	"google.golang.org/protobuf/proto"
)

var (
	shortNumberRegionToMetadataMap = make(map[string]*PhoneMetadata)
)

func readFromShortNumberRegionToMetadataMap(key string) (*PhoneMetadata, bool) {
	v, ok := shortNumberRegionToMetadataMap[key]
	return v, ok
}

func writeToShortNumberRegionToMetadataMap(key string, val *PhoneMetadata) {
	shortNumberRegionToMetadataMap[key] = val
}

func init() {
	err := loadShortNumberMetadataFromFile()
	if err != nil {
		panic(err)
	}
}

var (
	currShortNumberMetadataColl *PhoneMetadataCollection
	shortNumberReloadMetadata   = true
)

func ShortNumberMetadataCollection() (*PhoneMetadataCollection, error) {
	if !shortNumberReloadMetadata {
		return currShortNumberMetadataColl, nil
	}

	rawBytes, err := serialize.DecodeUnzip(shortNumberData)
	if err != nil {
		return nil, err
	}

	metadataCollection := &PhoneMetadataCollection{}
	err = proto.Unmarshal(rawBytes, metadataCollection)
	shortNumberReloadMetadata = false
	return metadataCollection, err
}

func loadShortNumberMetadataFromFile() error {
	metadataCollection, err := ShortNumberMetadataCollection()
	if err != nil {
		return err
	} else if currShortNumberMetadataColl == nil {
		currShortNumberMetadataColl = metadataCollection
	}

	metadataList := metadataCollection.GetMetadata()
	if len(metadataList) == 0 {
		return ErrEmptyMetadata
	}

	for _, meta := range metadataList {
		region := meta.GetId()
		if region == "001" {
			// it's a non geographical entity, unused
		} else {
			writeToShortNumberRegionToMetadataMap(region, meta)
		}
	}
	return nil
}

// ShortNumberCost is the cost category of a short number.
type ShortNumberCost int

// Cost categories of short numbers. Note these are suffixed with _COST to avoid
// clashing with the PhoneNumberType constants TOLL_FREE and PREMIUM_RATE.
const (
	TOLL_FREE_COST ShortNumberCost = iota
	STANDARD_RATE_COST
	PREMIUM_RATE_COST
	UNKNOWN_COST
)

func (c ShortNumberCost) String() string {
	switch c {
	case TOLL_FREE_COST:
		return "TOLL_FREE"
	case STANDARD_RATE_COST:
		return "STANDARD_RATE"
	case PREMIUM_RATE_COST:
		return "PREMIUM_RATE"
	default:
		return "UNKNOWN_COST"
	}
}

// Check whether a short number is a possible number. If a country calling code is shared by
// multiple regions, this returns true if it's possible in any of them. This provides a more
// lenient check than #isValidShortNumber.
// See IsPossibleShortNumberForRegion(PhoneNumber, string) for details.
func IsPossibleShortNumber(number *PhoneNumber) bool {
	regionsCodes := GetRegionCodesForCountryCode(int(number.GetCountryCode()))
	shortNumberLength := len(GetNationalSignificantNumber(number))
	for _, region := range regionsCodes {
		phoneMetadata := getShortNumberMetadataForRegion(region)
		if phoneMetadata == nil {
			continue
		}
		if phoneMetadata.GeneralDesc.HasPossibleLength(int32(shortNumberLength)) {
			return true
		}
	}
	return false
}

// Check whether a short number is a possible number when dialed from the given region. This
// provides a more lenient check than IsValidShortNumberForRegion.
func IsPossibleShortNumberForRegion(number *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(number, regionDialingFrom) {
		return false
	}
	phoneMetadata := getShortNumberMetadataForRegion(regionDialingFrom)
	if phoneMetadata == nil {
		return false
	}
	numberLength := len(GetNationalSignificantNumber(number))
	return phoneMetadata.GeneralDesc.HasPossibleLength(int32(numberLength))
}

// Tests whether a short number matches a valid pattern. If a country calling code is shared by
// multiple regions, this returns true if it's valid in any of them. Note that this doesn't verify
// the number is actually in use, which is impossible to tell by just looking at the number
// itself. See IsValidShortNumberForRegion(PhoneNumber, String) for details.
func IsValidShortNumber(number *PhoneNumber) bool {
	regionCodes := GetRegionCodesForCountryCode(int(number.GetCountryCode()))
	regionCode := getRegionCodeForShortNumberFromRegionList(number, regionCodes)
	if len(regionCodes) > 1 && regionCode != "" {
		// If a matching region had been found for the phone number from among two or more regions,
		// then we have already implicitly verified its validity for that region.
		return true
	}
	return IsValidShortNumberForRegion(number, regionCode)
}

// Tests whether a short number matches a valid pattern in a region. Note that this doesn't verify
// the number is actually in use, which is impossible to tell by just looking at the number itself.
func IsValidShortNumberForRegion(number *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(number, regionDialingFrom) {
		return false
	}
	phoneMetadata := getShortNumberMetadataForRegion(regionDialingFrom)
	if phoneMetadata == nil {
		return false
	}
	shortNumber := GetNationalSignificantNumber(number)
	generalDesc := phoneMetadata.GeneralDesc
	if !matchesPossibleNumberAndNationalNumber(shortNumber, generalDesc) {
		return false
	}
	shortNumberDesc := phoneMetadata.GetShortCode()
	return matchesPossibleNumberAndNationalNumber(shortNumber, shortNumberDesc)
}

// GetExpectedCostForRegion gets the expected cost category of a short number when dialed from a
// region (however, nothing is implied about its validity). If it is important that the number is
// valid, then its validity must first be checked using IsValidShortNumberForRegion. Note that
// emergency numbers are always considered toll-free. Returns UNKNOWN_COST if the number does not
// match a cost category. Note that an invalid number may match any cost category.
func GetExpectedCostForRegion(number *PhoneNumber, regionDialingFrom string) ShortNumberCost {
	if !regionDialingFromMatchesNumber(number, regionDialingFrom) {
		return UNKNOWN_COST
	}
	// Note that regionDialingFrom may be empty, in which case phoneMetadata will also be nil.
	phoneMetadata := getShortNumberMetadataForRegion(regionDialingFrom)
	if phoneMetadata == nil {
		return UNKNOWN_COST
	}

	shortNumber := GetNationalSignificantNumber(number)

	// The possible lengths are not present for a particular sub-type if they match the general
	// description; for this reason, we check the possible lengths against the general description
	// first to allow an early exit if possible.
	if !phoneMetadata.GetGeneralDesc().HasPossibleLength(int32(len(shortNumber))) {
		return UNKNOWN_COST
	}

	// The cost categories are tested in order of decreasing expense, since if for some reason the
	// patterns overlap the most expensive matching cost category should be returned.
	if matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.GetPremiumRate()) {
		return PREMIUM_RATE_COST
	}
	if matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.GetStandardRate()) {
		return STANDARD_RATE_COST
	}
	if matchesPossibleNumberAndNationalNumber(shortNumber, phoneMetadata.GetTollFree()) {
		return TOLL_FREE_COST
	}
	if IsEmergencyNumber(shortNumber, regionDialingFrom) {
		// Emergency numbers are implicitly toll-free.
		return TOLL_FREE_COST
	}
	return UNKNOWN_COST
}

// GetExpectedCost gets the expected cost category of a short number (however, nothing is implied
// about its validity). If the country calling code is unique to a region, this method behaves
// exactly the same as GetExpectedCostForRegion. However, if the country calling code is shared by
// multiple regions, then it returns the highest cost in the sequence PREMIUM_RATE, UNKNOWN_COST,
// STANDARD_RATE, TOLL_FREE. The reason for the position of UNKNOWN_COST in this order is that if a
// number is UNKNOWN_COST in one region but STANDARD_RATE or TOLL_FREE in another, its expected
// cost cannot be estimated as one of the latter since it might be a PREMIUM_RATE number.
//
// Note: If the region from which the number is dialed is known, it is highly preferable to call
// GetExpectedCostForRegion instead.
func GetExpectedCost(number *PhoneNumber) ShortNumberCost {
	regionCodes := GetRegionCodesForCountryCode(int(number.GetCountryCode()))
	if len(regionCodes) == 0 {
		return UNKNOWN_COST
	}
	if len(regionCodes) == 1 {
		return GetExpectedCostForRegion(number, regionCodes[0])
	}
	cost := TOLL_FREE_COST
	for _, regionCode := range regionCodes {
		costForRegion := GetExpectedCostForRegion(number, regionCode)
		switch costForRegion {
		case PREMIUM_RATE_COST:
			return PREMIUM_RATE_COST
		case UNKNOWN_COST:
			cost = UNKNOWN_COST
		case STANDARD_RATE_COST:
			if cost != UNKNOWN_COST {
				cost = STANDARD_RATE_COST
			}
		case TOLL_FREE_COST:
			// Do nothing.
		}
	}
	return cost
}

func getShortNumberMetadataForRegion(regionCode string) *PhoneMetadata {
	val, _ := readFromShortNumberRegionToMetadataMap(regionCode)
	return val
}

func getRegionCodeForShortNumberFromRegionList(number *PhoneNumber, regionCodes []string) string {
	if len(regionCodes) == 0 {
		return ""
	}
	if len(regionCodes) == 1 {
		return regionCodes[0]
	}
	nationalNumber := GetNationalSignificantNumber(number)
	for _, regionCode := range regionCodes {
		phoneMetadata := getShortNumberMetadataForRegion(regionCode)
		if phoneMetadata != nil && matchesPossibleNumberAndNationalNumber(nationalNumber, phoneMetadata.GetShortCode()) {
			// The number is valid for this region.
			return regionCode
		}
	}
	return ""
}

// getExampleShortNumber gets a valid short number for the specified region. Returns an empty
// string when the metadata does not contain such information.
func getExampleShortNumber(regionCode string) string {
	phoneMetadata := getShortNumberMetadataForRegion(regionCode)
	if phoneMetadata == nil {
		return ""
	}
	return phoneMetadata.GetShortCode().GetExampleNumber()
}

// getExampleShortNumberForCost gets a valid short number for the specified cost category. Returns
// an empty string when the metadata does not contain such information, or the cost is UNKNOWN_COST.
func getExampleShortNumberForCost(regionCode string, cost ShortNumberCost) string {
	phoneMetadata := getShortNumberMetadataForRegion(regionCode)
	if phoneMetadata == nil {
		return ""
	}
	var desc *PhoneNumberDesc
	switch cost {
	case TOLL_FREE_COST:
		desc = phoneMetadata.GetTollFree()
	case STANDARD_RATE_COST:
		desc = phoneMetadata.GetStandardRate()
	case PREMIUM_RATE_COST:
		desc = phoneMetadata.GetPremiumRate()
	default:
		// UNKNOWN_COST numbers are computed by the process of elimination from the other cost
		// categories.
	}
	if desc != nil {
		return desc.GetExampleNumber()
	}
	return ""
}

// Helper method to check that the country calling code of the number matches the region it's
// being dialed from.
func regionDialingFromMatchesNumber(number *PhoneNumber, regionDialingFrom string) bool {
	regionCodes := GetRegionCodesForCountryCode(int(number.GetCountryCode()))
	for _, region := range regionCodes {
		if region == regionDialingFrom {
			return true
		}
	}
	return false
}

// TODO: Once we have benchmarked ShortNumberInfo, consider if it is worth keeping
// this performance optimization.
func matchesPossibleNumberAndNationalNumber(number string, numberDesc *PhoneNumberDesc) bool {
	if numberDesc == nil {
		return false
	}
	if len(numberDesc.PossibleLength) > 0 && !numberDesc.HasPossibleLength(int32(len(number))) {
		return false
	}
	return regexbasedmatcher.MatchNationalNumber(number, numberDesc, false)
}

// In these countries, if extra digits are added to an emergency number, it no longer connects
// to the emergency service.
var regionsWhereEmergencyNumbersMustBeExact = []string{"BR", "CL", "NI"}

func matchesEmergencyNumber(number string, regionCode string, allowPrefixMatch bool) bool {
	possibleNumber := extractPossibleNumber(number)
	// Returns false if the number starts with a plus sign. We don't believe dialing the country
	// code before emergency numbers (e.g. +1911) works, but later, if that proves to work, we can
	// add additional logic here to handle it.
	if plusCharsPattern.MatchString(possibleNumber) {
		return false
	}

	phoneMetadata := getShortNumberMetadataForRegion(regionCode)
	if phoneMetadata == nil || phoneMetadata.GetEmergency() == nil {
		return false
	}

	normalizedNumber := NormalizeDigitsOnly(possibleNumber)

	allowPrefixMatchForRegion := allowPrefixMatch && !slices.Contains(regionsWhereEmergencyNumbersMustBeExact, regionCode)
	return regexbasedmatcher.MatchNationalNumber(normalizedNumber, phoneMetadata.GetEmergency(), allowPrefixMatchForRegion)
}

// Returns true if the given number exactly matches an emergency service number in the given
// region.
//
// This method takes into account cases where the number might contain formatting, but doesn't
// allow additional digits to be appended. Note that isEmergencyNumber(number, region)
// implies connectsToEmergencyNumber(number, region).
//
// number: the phone number to test
// regionCode: the region where the phone number is being dialed
// return: whether the number exactly matches an emergency services number in the given region
func IsEmergencyNumber(number string, regionCode string) bool {
	return matchesEmergencyNumber(number, regionCode, false)
}

// Returns true if the given number, exactly as dialed, might be used to connect to an emergency
// service in the given region.
//
// This method accepts a string, rather than a PhoneNumber, because it needs to distinguish
// cases such as "+1 911" and "911", where the former may not connect to an emergency service in
// all cases but the latter would. This method takes into account cases where the number might
// contain formatting, or might have additional digits appended (when it is okay to do that in
// the specified region).
//
// number: the phone number to test
// regionCode: the region where the phone number is being dialed
// return: whether the number might be used to connect to an emergency service in the given region
func ConnectsToEmergencyNumber(number string, regionCode string) bool {
	return matchesEmergencyNumber(number, regionCode, true)
}

// IsCarrierSpecific given a valid short number, determines whether it is carrier-specific (however,
// nothing is implied about its validity). Carrier-specific numbers may connect to a different
// end-point, or not connect at all, depending on the user's carrier. If it is important that the
// number is valid, then its validity must first be checked using IsValidShortNumber or
// IsValidShortNumberForRegion.
func IsCarrierSpecific(number *PhoneNumber) bool {
	regionCodes := GetRegionCodesForCountryCode(int(number.GetCountryCode()))
	regionCode := getRegionCodeForShortNumberFromRegionList(number, regionCodes)
	nationalNumber := GetNationalSignificantNumber(number)
	phoneMetadata := getShortNumberMetadataForRegion(regionCode)
	return phoneMetadata != nil &&
		matchesPossibleNumberAndNationalNumber(nationalNumber, phoneMetadata.GetCarrierSpecific())
}

// IsCarrierSpecificForRegion given a valid short number, determines whether it is carrier-specific
// when dialed from the given region (however, nothing is implied about its validity). Returns false
// if the number doesn't match the region provided.
func IsCarrierSpecificForRegion(number *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(number, regionDialingFrom) {
		return false
	}
	nationalNumber := GetNationalSignificantNumber(number)
	phoneMetadata := getShortNumberMetadataForRegion(regionDialingFrom)
	return phoneMetadata != nil &&
		matchesPossibleNumberAndNationalNumber(nationalNumber, phoneMetadata.GetCarrierSpecific())
}

// IsSmsServiceForRegion given a valid short number, determines whether it is an SMS service
// (however, nothing is implied about its validity). An SMS service is where the primary or only
// intended usage is to receive and/or send text messages (SMSs). This includes MMS as MMS numbers
// downgrade to SMS if the other party isn't MMS-capable. Returns false if the number doesn't match
// the region provided.
func IsSmsServiceForRegion(number *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(number, regionDialingFrom) {
		return false
	}
	phoneMetadata := getShortNumberMetadataForRegion(regionDialingFrom)
	return phoneMetadata != nil &&
		matchesPossibleNumberAndNationalNumber(GetNationalSignificantNumber(number), phoneMetadata.GetSmsServices())
}
