// Package stringbuilder provides a minimal, mutable string buffer used
// internally by the libphonenumber port. It is internal because it only exists
// to mirror the slice of java.lang.StringBuilder the port relies on and is not
// part of the public API.
package stringbuilder

import (
	"bytes"
	"errors"
	"slices"
	"unicode/utf8"
)

// ErrInvalidIndex is returned by Insert/InsertString/ByteAt when the given
// index falls outside the buffer.
var ErrInvalidIndex = errors.New("stringbuilder: invalid index")

// Builder is a minimal, mutable string buffer that mirrors the parts of
// java.lang.StringBuilder the libphonenumber port relies on — chiefly the
// ability to insert at an arbitrary index, which neither strings.Builder nor
// bytes.Buffer provide. It is backed by a plain []byte and leans on the
// standard library (slices.Insert, utf8.AppendRune) for the heavy lifting.
type Builder struct {
	buf []byte
}

// New creates a Builder using buf as its initial contents.
func New(buf []byte) *Builder { return &Builder{buf: buf} }

// NewString creates a Builder using s as its initial contents.
func NewString(s string) *Builder { return &Builder{buf: []byte(s)} }

// Len returns the number of bytes in the buffer.
func (b *Builder) Len() int { return len(b.buf) }

// Bytes returns the underlying contents. The slice is valid until the next
// mutating call; callers that retain it should copy.
func (b *Builder) Bytes() []byte { return b.buf }

// String returns the contents as a string. A nil receiver returns "<nil>",
// which is handy when debugging.
func (b *Builder) String() string {
	if b == nil {
		return "<nil>"
	}
	return string(b.buf)
}

// Write appends p to the buffer. The error is always nil; it exists to satisfy
// io.Writer.
func (b *Builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteString appends s to the buffer. The error is always nil.
func (b *Builder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// WriteByte appends c to the buffer. The error is always nil; it exists to
// satisfy io.ByteWriter.
func (b *Builder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of r to the buffer and returns the
// number of bytes written. The error is always nil.
func (b *Builder) WriteRune(r rune) (int, error) {
	n := len(b.buf)
	b.buf = utf8.AppendRune(b.buf, r)
	return len(b.buf) - n, nil
}

// Insert inserts p at index i, growing the buffer as needed. It returns
// ErrInvalidIndex if i is negative or greater than the buffer length.
func (b *Builder) Insert(i int, p []byte) (int, error) {
	if i < 0 || i > len(b.buf) {
		return -1, ErrInvalidIndex
	}
	b.buf = slices.Insert(b.buf, i, p...)
	return len(p), nil
}

// InsertString inserts s at index i, growing the buffer as needed. It returns
// ErrInvalidIndex if i is negative or greater than the buffer length.
func (b *Builder) InsertString(i int, s string) (int, error) {
	return b.Insert(i, []byte(s))
}

// ByteAt returns the byte at index i, or ErrInvalidIndex if i is out of range.
func (b *Builder) ByteAt(i int) (byte, error) {
	if i < 0 || i >= len(b.buf) {
		return 0, ErrInvalidIndex
	}
	return b.buf[i], nil
}

// Reset empties the buffer while retaining its capacity.
func (b *Builder) Reset() { b.buf = b.buf[:0] }

// ResetWith replaces the buffer's contents with buf.
func (b *Builder) ResetWith(buf []byte) (int, error) {
	b.buf = append(b.buf[:0], buf...)
	return len(buf), nil
}

// ResetWithString replaces the buffer's contents with s.
func (b *Builder) ResetWithString(s string) (int, error) {
	b.buf = append(b.buf[:0], s...)
	return len(s), nil
}

// The methods below mirror the remaining java.lang.StringBuilder operations that
// AsYouTypeFormatter relies on. Like their Java counterparts they assume valid
// indices and will panic on out-of-range access rather than returning an error.

// CharAt returns the byte at index i, mirroring StringBuilder.charAt. Callers in
// the port only ever index ASCII content, so a byte is an exact analogue of
// Java's char here.
func (b *Builder) CharAt(i int) byte { return b.buf[i] }

// SetLength truncates the buffer to n bytes, or pads it with NUL bytes if n is
// greater than the current length, mirroring StringBuilder.setLength.
func (b *Builder) SetLength(n int) {
	if n <= len(b.buf) {
		b.buf = b.buf[:n]
		return
	}
	for len(b.buf) < n {
		b.buf = append(b.buf, 0)
	}
}

// Delete removes the bytes in [start, end), mirroring StringBuilder.delete.
func (b *Builder) Delete(start, end int) {
	b.buf = append(b.buf[:start], b.buf[end:]...)
}

// Substring returns the bytes in [start, end) as a string, mirroring
// StringBuilder.substring(start, end).
func (b *Builder) Substring(start, end int) string {
	return string(b.buf[start:end])
}

// LastIndexOf returns the byte index of the last occurrence of s, or -1 if
// absent, mirroring StringBuilder.lastIndexOf.
func (b *Builder) LastIndexOf(s string) int {
	return bytes.LastIndex(b.buf, []byte(s))
}
