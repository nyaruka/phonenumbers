package phonenumbers

import (
	"golang.org/x/exp/slices"

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

	rawBytes, err := decodeUnzip(shortNumberData)
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
		if phoneMetadata.GeneralDesc.hasPossibleLength(int32(shortNumberLength)) {
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
	return phoneMetadata.GeneralDesc.hasPossibleLength(int32(numberLength))
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
	if len(numberDesc.PossibleLength) > 0 && !numberDesc.hasPossibleLength(int32(len(number))) {
		return false
	}
	return MatchNationalNumber(number, numberDesc, false)
}

// In these countries, if extra digits are added to an emergency number, it no longer connects
// to the emergency service.
var REGIONS_WHERE_EMERGENCY_NUMBERS_MUST_BE_EXACT = []string{"BR", "CL", "NI"}

func matchesEmergencyNumber(number string, regionCode string, allowPrefixMatch bool) bool {
	possibleNumber := extractPossibleNumber(number)
	// Returns false if the number starts with a plus sign. We don't believe dialing the country
	// code before emergency numbers (e.g. +1911) works, but later, if that proves to work, we can
	// add additional logic here to handle it.
	if PLUS_CHARS_PATTERN.MatchString(possibleNumber) {
		return false
	}

	phoneMetadata := getShortNumberMetadataForRegion(regionCode)
	if phoneMetadata == nil || phoneMetadata.GetEmergency() == nil {
		return false
	}

	normalizedNumber := NormalizeDigitsOnly(possibleNumber)

	allowPrefixMatchForRegion := allowPrefixMatch && !slices.Contains(REGIONS_WHERE_EMERGENCY_NUMBERS_MUST_BE_EXACT, regionCode)
	return MatchNationalNumber(normalizedNumber, phoneMetadata.GetEmergency(), allowPrefixMatchForRegion)
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
