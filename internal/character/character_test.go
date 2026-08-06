package character

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

// decadStarts returns the first code point of every ten-code-point run in
// unicode.Nd, and asserts the block structure Digit relies on: each range is a
// whole number of decads, has stride 1, and is a maximal run (so its first
// code point really is a zero). A future Unicode version that broke this would
// fail here rather than silently mis-decoding a script.
func decadStarts(t *testing.T) []rune {
	t.Helper()

	var starts []rune
	add := func(lo, hi rune, stride uint32) {
		assert.EqualValues(t, 1, stride, "unicode.Nd range U+%04X-U+%04X is not contiguous", lo, hi)
		assert.Zero(t, (hi-lo+1)%10, "unicode.Nd range U+%04X-U+%04X is not a whole number of decads", lo, hi)
		assert.False(t, unicode.IsDigit(lo-1), "unicode.Nd range U+%04X-U+%04X is not maximal at its start", lo, hi)
		assert.False(t, unicode.IsDigit(hi+1), "unicode.Nd range U+%04X-U+%04X is not maximal at its end", lo, hi)
		for c := lo; c <= hi; c += 10 {
			starts = append(starts, c)
		}
	}
	for _, r := range unicode.Nd.R16 {
		add(rune(r.Lo), rune(r.Hi), uint32(r.Stride))
	}
	for _, r := range unicode.Nd.R32 {
		add(rune(r.Lo), rune(r.Hi), r.Stride)
	}
	return starts
}

// TestDigit sweeps the whole rune space rather than a sample: Digit must
// accept exactly what unicode.IsDigit accepts (the predicate upstream Java
// spells Character.digit(c, 10) != -1) and yield the right value for all ten
// code points of every decad.
func TestDigit(t *testing.T) {
	for _, start := range decadStarts(t) {
		for i := rune(0); i < 10; i++ {
			v, ok := Digit(start + i)
			assert.True(t, ok, "U+%04X not recognized as a digit", start+i)
			assert.Equal(t, '0'+i, v, "wrong value for U+%04X", start+i)
		}
	}
	for c := rune(0); c <= unicode.MaxRune; c++ {
		_, ok := Digit(c)
		assert.Equal(t, unicode.IsDigit(c), ok, "Digit disagrees with unicode.IsDigit at U+%04X", c)
	}
}
