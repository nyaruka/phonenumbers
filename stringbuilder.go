package phonenumbers

import (
	"errors"
	"slices"
	"unicode/utf8"
)

// ErrInvalidIndex is returned by Insert/InsertString/ByteAt when the given
// index falls outside the buffer.
var ErrInvalidIndex = errors.New("phonenumbers.StringBuilder: invalid index")

// StringBuilder is a minimal, mutable string buffer that mirrors the parts of
// java.lang.StringBuilder the libphonenumber port relies on — chiefly the
// ability to insert at an arbitrary index, which neither strings.Builder nor
// bytes.Buffer provide. It is backed by a plain []byte and leans on the
// standard library (slices.Insert, utf8.AppendRune) for the heavy lifting.
type StringBuilder struct {
	buf []byte
}

// NewStringBuilder creates a StringBuilder using buf as its initial contents.
func NewStringBuilder(buf []byte) *StringBuilder { return &StringBuilder{buf: buf} }

// NewStringBuilderString creates a StringBuilder using s as its initial contents.
func NewStringBuilderString(s string) *StringBuilder { return &StringBuilder{buf: []byte(s)} }

// Len returns the number of bytes in the buffer.
func (b *StringBuilder) Len() int { return len(b.buf) }

// Bytes returns the underlying contents. The slice is valid until the next
// mutating call; callers that retain it should copy.
func (b *StringBuilder) Bytes() []byte { return b.buf }

// String returns the contents as a string. A nil receiver returns "<nil>",
// which is handy when debugging.
func (b *StringBuilder) String() string {
	if b == nil {
		return "<nil>"
	}
	return string(b.buf)
}

// Write appends p to the buffer. The error is always nil; it exists to satisfy
// io.Writer.
func (b *StringBuilder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteString appends s to the buffer. The error is always nil.
func (b *StringBuilder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// WriteByte appends c to the buffer. The error is always nil; it exists to
// satisfy io.ByteWriter.
func (b *StringBuilder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of r to the buffer and returns the
// number of bytes written. The error is always nil.
func (b *StringBuilder) WriteRune(r rune) (int, error) {
	n := len(b.buf)
	b.buf = utf8.AppendRune(b.buf, r)
	return len(b.buf) - n, nil
}

// Insert inserts p at index i, growing the buffer as needed. It returns
// ErrInvalidIndex if i is negative or greater than the buffer length.
func (b *StringBuilder) Insert(i int, p []byte) (int, error) {
	if i < 0 || i > len(b.buf) {
		return -1, ErrInvalidIndex
	}
	b.buf = slices.Insert(b.buf, i, p...)
	return len(p), nil
}

// InsertString inserts s at index i, growing the buffer as needed. It returns
// ErrInvalidIndex if i is negative or greater than the buffer length.
func (b *StringBuilder) InsertString(i int, s string) (int, error) {
	return b.Insert(i, []byte(s))
}

// ByteAt returns the byte at index i, or ErrInvalidIndex if i is out of range.
func (b *StringBuilder) ByteAt(i int) (byte, error) {
	if i < 0 || i >= len(b.buf) {
		return 0, ErrInvalidIndex
	}
	return b.buf[i], nil
}

// Reset empties the buffer while retaining its capacity.
func (b *StringBuilder) Reset() { b.buf = b.buf[:0] }

// ResetWith replaces the buffer's contents with buf.
func (b *StringBuilder) ResetWith(buf []byte) (int, error) {
	b.buf = append(b.buf[:0], buf...)
	return len(buf), nil
}

// ResetWithString replaces the buffer's contents with s.
func (b *StringBuilder) ResetWithString(s string) (int, error) {
	b.buf = append(b.buf[:0], s...)
	return len(s), nil
}
