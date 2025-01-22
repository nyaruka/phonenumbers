package accompanyhealth

import (
	"context"
	"fmt"

	logger "github.com/Accompany-Health/ah-logger"
	stats "github.com/Accompany-Health/ah-stats"
	libphonenumber "github.com/Accompany-Health/libphonenumber2"
)

// Filter out non-number characters from input phone number string
func FormatPhoneNumberForDatabase(phoneNumber string) (string, error) {
	num, err := libphonenumber.Parse(phoneNumber, "US")
	if err != nil {
		stats.Stats().Increment("phone.number.parse_error")
		logger.L().Warn("Can't parse phone number", phoneNumber, err)
		return "", err
	}
	if !libphonenumber.IsValidNumber(num) {
		stats.Stats().Increment("phone.number.invalid")
		logger.L().Warn("Invalid phone number", phoneNumber, err)
		return "", fmt.Errorf("invalid phone number")
	}
	return libphonenumber.Format(num, libphonenumber.E164), nil
}

func IsValidPhoneNumber(ctx context.Context, phoneNumber string) bool {
	phone, err := libphonenumber.ParseAndKeepRawInput(phoneNumber, "US")
	if err != nil {
		logger.L().WarnCtx(ctx, "Can't parse phone number `"+phoneNumber+"`", "phoneNumber", phoneNumber, "error", err)
		return false
	}

	//Assume all nbumbers from US for now
	return libphonenumber.IsValidNumberForRegion(phone, "US")
}

// Check for equality between two phone numbers
func IsEqual(phoneOne string, phoneTwo string) bool {
	first, err1 := FormatPhoneNumberForDatabase(phoneOne)
	second, err2 := FormatPhoneNumberForDatabase(phoneTwo)
	if err1 != nil || err2 != nil {
		return phoneOne == phoneTwo
	}
	return first == second
}
