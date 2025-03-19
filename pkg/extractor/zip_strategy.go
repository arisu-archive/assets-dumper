package extractor

import (
	"archive/zip"
	"compress/flate"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"

	"github.com/arisu-archive/assets-dumper/pkg/zipcrypto"
)

func zipExtractor(ctx context.Context, inputPath string) (*Result, error) {
	zipReader, zipReaderErr := zip.OpenReader(inputPath)
	if zipReaderErr != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", zipReaderErr)
	}
	slog.DebugContext(ctx, "Creating key from file name of input path", "inputPath", inputPath)
	fileName := filepath.Base(inputPath)
	key := fbsutils.CreateZipPassword(fileName)
	slog.DebugContext(ctx, "Key created", "key", string(key))

	files := make([]ExtractedFile, 0, len(zipReader.File))
	for _, file := range zipReader.File {
		fr, openZipErr := file.OpenRaw()
		if openZipErr != nil {
			return nil, fmt.Errorf("failed to open file: %w", openZipErr)
		}

		zr, zipCryptoErr := zipcrypto.NewReader(file.FileHeader, fr, key)
		if zipCryptoErr != nil {
			return nil, fmt.Errorf("failed to create zipcrypto reader: %w", zipCryptoErr)
		}

		// All flatdata is encrypted.
		files = append(files, ExtractedFile{
			Name:        file.Name,
			Size:        int(file.FileHeader.UncompressedSize64),
			Reader:      flate.NewReader(zr),
			IsEncrypted: true,
		})
	}

	// Folder should be the name of the zip file without the extension
	folder := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return &Result{
		rc:     zipReader,
		Folder: folder,
		Files:  files,
	}, nil
}
