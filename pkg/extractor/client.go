package extractor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/arisu-archive/assets-dumper/pkg/decryption"
	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

type Client interface {
	Extract(ctx context.Context, inputPath, outputPath string) error
}

// Option is a functional option for the extractor client.
type Option func(*client)

// WithKey sets the SQLCipher database key for encrypted database extraction.
func WithKey(key []byte) Option {
	return func(c *client) {
		c.key = key
	}
}

type client struct {
	server    resourceapi.Server
	decryptor decryption.Client
	key       []byte
}

func New(server resourceapi.Server, opts ...Option) Client {
	return NewWithDecryptor(server, decryption.New(server), opts...)
}

func NewWithDecryptor(server resourceapi.Server, d decryption.Client, opts ...Option) Client {
	c := &client{
		server:    server,
		decryptor: d,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *client) Extract(ctx context.Context, inputPath, outputPath string) error {
	// Determine the input path is a file or a directory
	files, err := getFiles(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get files: %w", err)
	}

	// Process each file based on the file extension
	for _, file := range files {
		extractor := c.getExtractor(FileFormat(filepath.Ext(file)))
		if extractor == nil {
			slog.WarnContext(
				ctx,
				"failed to get extractor. skipping file",
				"file", file,
			)
			continue
		}

		// Extract the file to the output path
		result, extractErr := extractor(ctx, file)
		if extractErr != nil {
			return fmt.Errorf("failed to extract file: %w", extractErr)
		}
		// Wrap the result with the decryption reader
		saveErr := result.WithDecryptor(c.decryptor).Save(ctx, outputPath)
		closeErr := result.Close()
		if saveErr != nil {
			return fmt.Errorf("failed to save file: %w", saveErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close file: %w", closeErr)
		}
	}

	return nil
}

// getExtractor returns the appropriate extractor for the given format.
// For Japan SQLite databases with a key, it returns a key-aware extractor.
func (c *client) getExtractor(format FileFormat) Extractor {
	if len(c.key) > 0 && format == fileFormatSqlite {
		return sqliteExtractorWithKey(c.server, c.key)
	}
	return extractor(c.server, format)
}

func getFiles(inputPath string) ([]string, error) {
	// Check if the input path is a file or a directory
	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get stat of %s: %w", inputPath, err)
	}

	if info.IsDir() {
		return getFilesFromDirectory(inputPath)
	}
	return []string{inputPath}, nil
}

func getFilesFromDirectory(inputPath string) ([]string, error) {
	// Recursively get all files in the directory
	// Use walk to get all files
	var files []string
	err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", inputPath, err)
	}

	return files, nil
}
