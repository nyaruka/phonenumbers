// Package character provides stand-ins for the slice of java.lang.Character
// the libphonenumber port relies on. It is internal because it only exists to
// mirror Java behaviour and is not part of the public API.
package character

import "unicode"

// Digit returns the ASCII digit for the Unicode decimal digit c, standing in
// for Java's Character.digit(c, 10). Every Nd block is ten consecutive code
// points in ascending value order, so a code point's offset within its block is
// its value. unicode.Nd is the table unicode.IsDigit itself consults, so this
// stays complete as Go's Unicode version moves.
//
// Over the BMP this matches Character.digit(c, 10) exactly. It deliberately
// also covers supplementary-plane Nd digits (e.g. Adlam, Osmanya, mathematical
// digits), which Java's per-char UTF-16 iteration hands to Character.digit as
// bare surrogates and so can never accept. The C++ port (u_charDigitValue over
// code points) and the Python port accept them too; we follow the ports rather
// than the encoding artifact. Recorded as a deliberate divergence in SYNC.md.
func Digit(c rune) (rune, bool) {
	if '0' <= c && c <= '9' {
		return c, true
	}
	for _, r := range unicode.Nd.R16 {
		if c < rune(r.Lo) {
			return 0, false
		}
		if c <= rune(r.Hi) {
			return '0' + (c-rune(r.Lo))%10, true
		}
	}
	for _, r := range unicode.Nd.R32 {
		if c < rune(r.Lo) {
			break
		}
		if c <= rune(r.Hi) {
			return '0' + (c-rune(r.Lo))%10, true
		}
	}
	return 0, false
}
