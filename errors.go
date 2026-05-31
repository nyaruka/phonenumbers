// Port of NumberParseException.java + package errors from google/libphonenumber.
// Functions are kept in upstream source order to ease syncing.
// Last reconciled against: v9.0.32
package phonenumbers

import "errors"

var ErrEmptyMetadata = errors.New("empty metadata")

var ErrTooShortAfterIDD = errors.New("phone number had an IDD, but " +
	"after this was not long enough to be a viable phone number")

var ErrNumTooLong = errors.New("the string supplied is too long to be a phone number")
