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

	"github.com/arisu-archive/arona-flatbuffers/go/flatdata"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

func flatdataReader(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error) {
	t := flatdata.GetFlatDataByName(strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name))))
	if t == nil {
		slog.WarnContext(ctx, "failed to get table by string", "name", name)
		return r, nil
	}
	slog.DebugContext(ctx, "Decrypting flatbuffer", "name", name)
	// Wrap it with our decryptor reader
	xr := newXorReader(r, fbsutils.CreateKey(t.FlatDataName(), size))
	// Decrypt the flatbuffer first using the key the flatbuffer takes []byte
	data, err := io.ReadAll(xr)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	slog.DebugContext(ctx, "Decrypted flatbuffer", "name", name, "size", len(data))

	return convertToJSON(ctx, t, data)
}

func convertToJSON(ctx context.Context, t fbsutils.FlatData, data []byte) (io.Reader, error) {
	if err := t.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flatbuffer: %w", err)
	}
	slog.DebugContext(ctx, "Unmarshalled flatbuffer", "name", t.FlatDataName(), "size", len(data))
	jsonData, err := json.MarshalIndent(&t, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flatbuffer: %w", err)
	}
	slog.DebugContext(ctx, "Marshaled flatbuffer", "name", t.FlatDataName(), "size", len(jsonData))
	return bytes.NewReader(jsonData), nil
}
