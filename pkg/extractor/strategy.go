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
		if err := e.saveFile(ctx, outputPath, f); err != nil {
			return err
		}
	}
	return nil
}

func (e *Result) saveFile(ctx context.Context, outputPath string, f ExtractedFile) error {
	slog.DebugContext(
		ctx,
		"Saving file",
		"file", f.Name,
		"outputPath", outputPath,
		"fullPath", filepath.Join(outputPath, f.Name),
	)

	reader, skip := e.prepareReader(ctx, f)
	if skip {
		return nil
	}

	if mkdirErr := os.MkdirAll(outputPath, 0o755); mkdirErr != nil {
		return fmt.Errorf("failed to create directory: %w", mkdirErr)
	}

	fileName := strings.TrimSuffix(f.Name, filepath.Ext(f.Name)) + ".json"
	return e.writeFile(filepath.Join(outputPath, fileName), reader)
}

func (e *Result) prepareReader(ctx context.Context, f ExtractedFile) (io.Reader, bool) {
	if e.decryptor == nil || !f.IsEncrypted {
		return f.Reader, false
	}

	dr, dErr := e.decryptor.DecryptionReader(ctx, f.Name, f.Size, f.Reader)
	if dErr != nil {
		if errors.Is(dErr, decryption.ErrFlatbufferUnmarshalFailed) {
			slog.WarnContext(
				ctx,
				"failed to unmarshal flatbuffer, skipping this file.",
				"file", f.Name, "error", dErr,
			)
			return nil, true
		}
		slog.WarnContext(ctx, "failed to setup decryption reader", "error", dErr)
		return f.Reader, false
	}
	return dr, false
}

func (*Result) writeFile(filePath string, reader io.Reader) error {
	file, createErr := os.Create(filePath)
	if createErr != nil {
		return fmt.Errorf("failed to create file: %w", createErr)
	}
	defer file.Close()

	if _, copyErr := io.Copy(file, reader); copyErr != nil && !errors.Is(copyErr, io.EOF) {
		return fmt.Errorf("failed to copy file: %w", copyErr)
	}

	if closeErr := file.Close(); closeErr != nil {
		return fmt.Errorf("failed to close file: %w", closeErr)
	}

	return nil
}
