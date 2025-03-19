package zipcrypto

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
)

var (
	ErrNotEncrypted    = errors.New("zip file is not encrypted")
	ErrInvalidPassword = errors.New("invalid password")
)

type zipCryptoReader struct {
	rr   io.Reader
	keys keys
}

func NewReader(fh zip.FileHeader, rc io.Reader, password []byte) (io.Reader, error) {
	// If it is not encrypted, return the reader
	if !IsEncrypted(fh.Flags) {
		return nil, ErrNotEncrypted
	}

	r := &zipCryptoReader{
		rr:   rc,
		keys: newKeys(password),
	}

	// Verify the password by reading the encryption header
	if verifyErr := r.verifyPassword(fh); verifyErr != nil {
		return nil, fmt.Errorf("failed to verify password: %w", verifyErr)
	}

	return r, nil
}

func (r *zipCryptoReader) Read(p []byte) (int, error) {
	n, readErr := r.rr.Read(p)
	if readErr != nil {
		return n, fmt.Errorf("failed to read file: %w", readErr)
	}
	for i := range p {
		p[i] = r.decrypt(p[i])
	}
	return n, nil
}

// verifyPassword checks if the password is correct by decrypting the 12-byte header.
func (r *zipCryptoReader) verifyPassword(header zip.FileHeader) error {
	// Read the 12-byte encryption header
	headerBytes := make([]byte, 12)
	if _, readErr := io.ReadFull(r.rr, headerBytes); readErr != nil {
		return fmt.Errorf("failed to read encryption header: %w", readErr)
	}

	// Decrypt the header
	for i := range headerBytes {
		headerBytes[i] = r.decrypt(headerBytes[i])
	}

	//nolint:staticcheck // ZipCrypto uses this field to verify the password
	expectedByte := byte(header.ModifiedTime >> 8)
	// Does not have data descriptor
	if !HasDataDescriptor(header.Flags) {
		expectedByte = byte(header.CRC32 >> 24)
	}

	if headerBytes[11] != expectedByte {
		return ErrInvalidPassword
	}

	return nil
}

func (r *zipCryptoReader) decrypt(b byte) byte {
	b ^= r.keys.Byte()
	r.keys.update(b)
	return b
}
