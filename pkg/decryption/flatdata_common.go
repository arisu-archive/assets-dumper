package decryption

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

var ErrFlatbufferUnmarshalFailed = errors.New("failed to unmarshal flatbuffer")

// FlatDataProvider defines the interface for getting flatdata by name
type FlatDataProvider interface {
	GetFlatDataByName(name string) fbsutils.FlatData
}

func flatdataReaderCommon(ctx context.Context, provider FlatDataProvider, name string, size uint64, r io.Reader) (io.Reader, error) {
	t := provider.GetFlatDataByName(strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name))))
	if t == nil {
		slog.WarnContext(ctx, "failed to get table by string", "name", name)
		return r, nil
	}
	slog.DebugContext(ctx, "Decrypting flatbuffer", "name", name)
	// Wrap it with our decryptor reader
	xr := NewXorReader(r, fbsutils.CreateKey(t.FlatDataName(), size))
	// Decrypt the flatbuffer first using the key the flatbuffer takes []byte
	data, err := io.ReadAll(xr)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	slog.DebugContext(ctx, "Decrypted flatbuffer", "name", name, "size", len(data))

	jsonReader, err := convertToJSON(ctx, t, data)
	if err != nil {
		slog.WarnContext(ctx, "failed to convert to JSON", "name", name, "error", err)
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}
	return jsonReader, nil
}

func convertToJSON(ctx context.Context, t fbsutils.FlatData, data []byte) (out io.Reader, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.WarnContext(ctx, "panic while unmarshalling flatbuffer", "name", t.FlatDataName(), "error", r)
			err = fmt.Errorf("%w: %v", ErrFlatbufferUnmarshalFailed, r)
			out = nil
		}
	}()
	if err := t.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFlatbufferUnmarshalFailed, err)
	}
	slog.DebugContext(ctx, "Unmarshalled flatbuffer", "name", t.FlatDataName(), "size", len(data))
	jsonData, err := json.MarshalIndent(&t, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flatbuffer: %w", err)
	}
	slog.DebugContext(ctx, "Marshaled flatbuffer", "name", t.FlatDataName(), "size", len(jsonData))
	return bytes.NewReader(jsonData), nil
}
