package zipcrypto

import "hash/crc32"

type keys [3]uint32

const (
	// https://pkwaredownloads.blob.core.windows.net/pkware-general/Documentation/APPNOTE-6.2.0.txt
	Keys0 = 0x12345678 // 305419896
	Keys1 = 0x23456789 // 590708185
	Keys2 = 0x34567890 // 875996472
)

func newKeys(password []byte) keys {
	k := keys{Keys0, Keys1, Keys2}
	for _, p := range password {
		k.update(p)
	}
	return k
}

func (k *keys) update(p byte) {
	k[0] = updateCrc32(k[0], p)
	k[1] += k[0] & 0xff
	k[1] = k[1]*134775813 + 1
	k[2] = updateCrc32(k[2], byte(k[1]>>24))
}

func (k *keys) Byte() byte {
	t := k[2] | 2
	return byte((t * (t ^ 1)) >> 8)
}

func updateCrc32(crc uint32, b byte) uint32 {
	return crc32.IEEETable[(crc^uint32(b))&0xff] ^ (crc >> 8)
}
