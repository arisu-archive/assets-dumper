package decryption

import (
	"fmt"
	"io"
)

type reader struct {
	r   io.Reader
	pos int
	key []byte
}

func NewXorReader(r io.Reader, key []byte) io.Reader {
	return &reader{r: r, key: key}
}

func (r *reader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	keyOffset := r.pos % len(r.key)
	for i := range n {
		p[i] ^= r.key[(keyOffset+i)%len(r.key)]
	}
	r.pos += n
	if err != nil {
		return n, fmt.Errorf("failed to read: %w", err)
	}

	return n, nil
}
