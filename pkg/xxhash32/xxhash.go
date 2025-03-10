package xxhash32

import (
	"errors"
	"hash"
)

var ErrInputTooLarge = errors.New("xxhash: input too large")

const (
	Prime1 = uint32(0x9E3779B1) // 0b10011110001101110111100110110001
	Prime2 = uint32(0x85EBCA77) // 0b10000101111010111100101001110111
	Prime3 = uint32(0xC2B2AE3D) // 0b11000010101100101010111000111101
	Prime4 = uint32(0x27D4EB2F) // 0b00100111110101001110101000101111
	Prime5 = uint32(0x165667B1) // 0b00010110010101100110011110110001
)

type xxHash struct {
	seed  uint32
	acc1  uint32
	acc2  uint32
	acc3  uint32
	acc4  uint32
	buf   [16]byte
	total uint32
	n     int // number of bytes in buf
}

// New returns a new hash.Hash32 that uses the default seed value.
func New() hash.Hash32 {
	return NewWithSeed(0)
}

// NewWithSeed returns a new hash.Hash32 that uses the given seed value.
func NewWithSeed(seed uint32) hash.Hash32 {
	h := &xxHash{}
	return h.WithSeed(seed)
}

// WithSeed returns a new hash.Hash32 that uses the given seed value.
func (h *xxHash) WithSeed(seed uint32) hash.Hash32 {
	h.seed = seed
	h.Reset()
	return h
}

// Reset resets the hash to its initial state.
func (h *xxHash) Reset() {
	h.acc1 = h.seed + Prime1 + Prime2
	h.acc2 = h.seed + Prime2
	h.acc3 = h.seed + 0
	h.acc4 = h.seed - Prime1
	h.buf = [16]byte{}
	h.total = 0
	h.n = 0
}

// Size returns the size of the hash in bytes.
func (*xxHash) Size() int {
	return 4
}

// BlockSize returns the hash's underlying block size.
func (*xxHash) BlockSize() int {
	return 16
}

// Write adds input to the hash and returns the number of bytes written.
func (h *xxHash) Write(input []byte) (int, error) {
	needed := len(input)
	// Check if needed is too large for uint32
	if needed > int(^uint32(0)) {
		return 0, ErrInputTooLarge
	}
	h.total += uint32(needed)

	// Does not have enough data to fill the current block
	remained := len(h.buf) - h.n
	if needed < remained {
		copy(h.buf[h.n:], input)
		h.n += needed
		return needed, nil
	}

	// Finish the current block
	if h.n > 0 {
		// Copy the input into the remaining part of the buffer
		c := copy(h.buf[h.n:], input)
		h.acc1 = round(h.acc1, u32(h.buf[0:4]))
		h.acc2 = round(h.acc2, u32(h.buf[4:8]))
		h.acc3 = round(h.acc3, u32(h.buf[8:12]))
		h.acc4 = round(h.acc4, u32(h.buf[12:16]))
		input = input[c:]
		h.n = 0
	}

	// Process the remaining input
	for len(input) >= 16 {
		r := h.processBlock(input)
		input = input[r:]
	}

	// Copy the remaining input into the buffer
	copy(h.buf[:], input)
	h.n = len(input)

	return needed, nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the state of the hash.
func (h *xxHash) Sum(b []byte) []byte {
	c := h.Sum32()
	return append(b, byte(c>>24), byte(c>>16), byte(c>>8), byte(c))
}

// Sum32 returns the current hash as a uint32.
func (h *xxHash) Sum32() uint32 {
	acc := h.total
	if h.total < 16 {
		acc += h.acc3 + Prime5
	} else {
		acc += rol1(h.acc1) + rol7(h.acc2) + rol12(h.acc3) + rol18(h.acc4)
	}

	// Consume the remaining input in the buffer
	p := 0
	// Process 4-byte chunks
	for p+4 <= h.n {
		acc += u32(h.buf[p:p+4]) * Prime3
		acc = rol17(acc) * Prime4
		p += 4
	}
	// Process remaining bytes
	for p < h.n {
		acc += uint32(h.buf[p]) * Prime5
		acc = rol11(acc) * Prime1
		p++
	}

	acc ^= acc >> 15
	acc *= Prime2
	acc ^= acc >> 13
	acc *= Prime3
	return acc ^ (acc >> 16)
}

func Checksum(b []byte) uint32 {
	h := New()
	h.Write(b)
	return h.Sum32()
}

func (h *xxHash) processBlock(input []byte) int {
	v1, v2, v3, v4 := h.acc1, h.acc2, h.acc3, h.acc4
	n := len(input)
	for len(input) >= 16 {
		v1 = round(v1, u32(input[0:4:len(input)]))
		v2 = round(v2, u32(input[4:8:len(input)]))
		v3 = round(v3, u32(input[8:12:len(input)]))
		v4 = round(v4, u32(input[12:16:len(input)]))
		input = input[16:len(input):len(input)]
	}
	h.acc1, h.acc2, h.acc3, h.acc4 = v1, v2, v3, v4
	// Report the remaining bytes
	return n - len(input)
}
