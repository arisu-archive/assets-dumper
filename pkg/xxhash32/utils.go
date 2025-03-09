package xxhash32

import "math/bits"

func u32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func round(acc, b uint32) uint32 {
	acc += b * Prime2
	acc = rol13(acc) * Prime1
	return acc
}

func rol1(v uint32) uint32  { return bits.RotateLeft32(v, 1) }
func rol7(v uint32) uint32  { return bits.RotateLeft32(v, 7) }
func rol11(v uint32) uint32 { return bits.RotateLeft32(v, 11) }
func rol12(v uint32) uint32 { return bits.RotateLeft32(v, 12) }
func rol13(v uint32) uint32 { return bits.RotateLeft32(v, 13) }
func rol17(v uint32) uint32 { return bits.RotateLeft32(v, 17) }
func rol18(v uint32) uint32 { return bits.RotateLeft32(v, 18) }
