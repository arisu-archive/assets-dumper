package extractor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/arisu-archive/assets-dumper/pkg/decryption"
)

type Result struct {
	rc        io.Closer
	Folder    string
	Files     []ExtractedFile
	decryptor decryption.Client
}

type ExtractedFile struct {
	Name        string
	Size        uint64
	Reader      io.Reader
	IsEncrypted bool
}

func (e *Result) Close() error {
	if err := e.rc.Close(); err != nil {
		return fmt.Errorf("failed to close zip reader: %w", err)
	}

	return nil
}

func (e *Result) WithDecryptor(d decryption.Client) *Result {
	e.decryptor = d
	return e
}

func (e *Result) Save(ctx context.Context, outputDir string) error {
	outputPath := filepath.Join(outputDir, e.Folder)
	for _, f := range e.Files {
		slog.DebugContext(
			ctx,
			"Saving file",
			"file", f.Name,
			"outputPath", outputPath,
			"fullPath", filepath.Join(outputPath, f.Name),
		)
		extension := ".json"
		if e.decryptor != nil && f.IsEncrypted {
			// We need to use the flatdata struct name to derive the key
			dr, dErr := e.decryptor.DecryptionReader(ctx, f.Name, f.Size, f.Reader)
			if dErr != nil {
				if errors.Is(dErr, decryption.ErrFlatbufferUnmarshalFailed) {
					slog.WarnContext(ctx, "failed to unmarshal flatbuffer, skipping this file.", "file", f.Name, "error", dErr)
					continue
				}
				slog.WarnContext(ctx, "failed to setup decryption reader", "error", dErr)
				goto normal
			}
			f.Reader = dr
		}
	normal:
		if mkdirErr := os.MkdirAll(outputPath, 0o755); mkdirErr != nil {
			return fmt.Errorf("failed to create directory: %w", mkdirErr)
		}
		fileName := strings.TrimSuffix(f.Name, filepath.Ext(f.Name)) + extension
		file, createErr := os.Create(filepath.Join(outputPath, fileName))
		if createErr != nil {
			return fmt.Errorf("failed to create file: %w", createErr)
		}

		if _, copyErr := io.Copy(file, f.Reader); copyErr != nil && !errors.Is(copyErr, io.EOF) {
			return fmt.Errorf("failed to copy file: %w", copyErr)
		}

		if closeErr := file.Close(); closeErr != nil {
			return fmt.Errorf("failed to close file: %w", closeErr)
		}
	}

	return nil
}
