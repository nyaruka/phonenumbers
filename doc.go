// Package phonenumbers is a Go port of Google's libphonenumber for parsing,
// formatting, and validating international phone numbers.
//
// It tracks the Java implementation — libphonenumber's reference
// implementation, of which the C++ and JavaScript versions are themselves
// ports. As a result, type names, method names, and tests deliberately mirror
// their Java counterparts (for example PhoneNumberUtil and PhoneNumberUtilTest)
// to keep the port easy to verify against upstream. The aim is strictly to
// match libphonenumber's functionality rather than to add to it.
//
// See SYNC.md for which upstream version each ported file is reconciled against.
package phonenumbers
