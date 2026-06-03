// Port of java/libphonenumber/src/com/google/i18n/phonenumbers/NumberParseException.java (+ package errors).
package phonenumbers

import (
	"errors"

	"github.com/nyaruka/phonenumbers/v2/metadata"
)

// ErrEmptyMetadata is an alias for metadata.ErrEmptyMetadata, kept here for
// backwards compatibility.
var ErrEmptyMetadata = metadata.ErrEmptyMetadata

var ErrTooShortAfterIDD = errors.New("phone number had an IDD, but " +
	"after this was not long enough to be a viable phone number")

var ErrNumTooLong = errors.New("the string supplied is too long to be a phone number")
