// Copyright 2017
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package starlark

// Bits represents a fixed-size bit string.
// Its zero value is an empty bit string.
type Bits struct {
	bytes []byte
}

// NewBits returns a new bit string of length n bits.
func NewBits(n int) *Bits {
	if n < 0 {
		panic("negative length")
	}
	return &Bits{bytes: make([]byte, (n+7)/8)}
}

// Set sets the bit i.
func (b *Bits) Set(i int) {
	b.bytes[i>>3] |= 1 << uint(i&7)
}

// Clear clears the bit i.
func (b *Bits) Clear(i int) {
	b.bytes[i>>3] &^= 1 << uint(i&7)
}

// Reset clears all bits in the bit string.
func (b *Bits) Reset() {
	for i := range b.bytes {
		b.bytes[i] = 0
	}
}

// Get returns the value of bit i.
func (b *Bits) Get(i int) bool {
	return b.bytes[i>>3]&(1<<uint(i&7)) != 0
}
