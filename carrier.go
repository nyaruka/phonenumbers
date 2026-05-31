// Port of carrier/PhoneNumberToCarrierMapper.java from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
// Last reconciled against: v9.0.32
package phonenumbers

// GetCarrierForNumber returns the carrier we believe the number belongs to. Note due
// to number porting this is only a guess, there is no guarantee to its accuracy.
func GetCarrierForNumber(number *PhoneNumber, lang string) (string, error) {
	carrier, _, err := GetCarrierWithPrefixForNumber(number, lang)
	return carrier, err
}

// GetSafeCarrierDisplayNameForNumber Gets the name of the carrier for the given phone number
// only when it is 'safe' to display to users.
// A carrier name is considered safe if the number is valid and
// for a region that doesn't support mobile number portability .
func GetSafeCarrierDisplayNameForNumber(phoneNumber *PhoneNumber, lang string) (string, error) {
	if IsMobileNumberPortableRegion(GetRegionCodeForNumber(phoneNumber)) {
		return "", nil
	}
	return GetCarrierForNumber(phoneNumber, lang)
}

// GetCarrierWithPrefixForNumber returns the carrier we believe the number belongs to, as well as
// its prefix. Note due to number porting this is only a guess, there is no guarantee to its accuracy.
func GetCarrierWithPrefixForNumber(number *PhoneNumber, lang string) (string, int, error) {
	carrier, prefix, err := getValueForNumber(carrierPrefixMap, carrierData, carrierDataPath, lang, 10, number)
	if err != nil {
		return "", 0, err
	}
	if carrier != "" {
		return carrier, prefix, nil
	}

	// fallback to english
	return getValueForNumber(carrierPrefixMap, carrierData, carrierDataPath, "en", 10, number)
}
