package extractor

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	_ "github.com/mutecomm/go-sqlcipher/v4" // sqlite3 driver with sqlcipher support.
)

func sqliteExtractorCommonWithKey(ctx context.Context, provider ExcelProvider, inputPath string, key []byte) (*Result, error) {
	db, err := openSqlcipherDatabase(inputPath, key)
	if err != nil {
		return nil, err
	}

	return sqliteExtractorFromDB(ctx, provider, db, inputPath)
}

func sqliteExtractorCommon(ctx context.Context, provider ExcelProvider, inputPath string) (*Result, error) {
	db, err := openSqliteDatabase(inputPath)
	if err != nil {
		return nil, err
	}

	return sqliteExtractorFromDB(ctx, provider, db, inputPath)
}

func sqliteExtractorWithKey(key []byte) Extractor {
	return func(ctx context.Context, inputPath string) (*Result, error) {
		return sqliteExtractorCommonWithKey(ctx, japanExcelProvider, inputPath, key)
	}
}

func sqliteExtractorFromDB(ctx context.Context, provider ExcelProvider, db *sql.DB, inputPath string) (*Result, error) {
	defer db.Close()

	tableNames, err := getSqliteTableNames(ctx, db)
	if err != nil {
		return nil, err
	}

	files, err := processAllTablesCommon(ctx, provider, db, tableNames)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Base(inputPath)
	folder := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return &Result{
		rc:     db,
		Folder: folder,
		Files:  files,
	}, nil
}

func openSqliteDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}
	return db, nil
}

func openSqlcipherDatabase(path string, key []byte) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_pragma_key=x'%x'&_pragma_cipher_page_size=4096", path, key)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Verify the key is correct by reading the database.
	if _, queryErr := db.Exec("SELECT count(*) FROM sqlite_master"); queryErr != nil {
		db.Close()
		return nil, fmt.Errorf("failed to verify sqlcipher key (wrong key?): %w", queryErr)
	}

	return db, nil
}

func getSqliteTableNames(ctx context.Context, db *sql.DB) ([]string, error) {
	tblRows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer tblRows.Close()

	var tableNames []string
	for tblRows.Next() {
		var name string
		if scanErr := tblRows.Scan(&name); scanErr != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", scanErr)
		}
		tableNames = append(tableNames, name)
	}

	if tblRows.Err() != nil {
		return nil, fmt.Errorf("failed to get tables: %w", tblRows.Err())
	}

	for _, name := range tableNames {
		slog.DebugContext(ctx, "Found table", "name", name)
	}

	return tableNames, nil
}

func processAllTablesCommon(
	ctx context.Context,
	provider ExcelProvider,
	db *sql.DB,
	tableNames []string,
) ([]ExtractedFile, error) {
	files := make([]ExtractedFile, 0, len(tableNames))

	for _, name := range tableNames {
		file, err := processTableCommon(ctx, provider, db, name)
		if err != nil {
			return nil, err
		}

		if file != nil {
			files = append(files, *file)
		}
	}

	return files, nil
}

func processTableCommon(
	ctx context.Context,
	provider ExcelProvider,
	db *sql.DB,
	tableName string,
) (*ExtractedFile, error) {
	slog.DebugContext(ctx, "Table", "name", strings.TrimSuffix(tableName, "DBSchema"))

	// Some tables contains underscore, and our mapping is from the actual struct name.
	flatbufferName := strings.ReplaceAll(strings.TrimSuffix(tableName, "DBSchema")+"Excel", "_", "")

	// Query the table
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT Bytes FROM %s", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}
	defer rows.Close()

	schemaData, err := extractTableDataCommon(ctx, provider, rows, flatbufferName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(map[string]any{
		"data_list": schemaData,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal row: %w", err)
	}

	return &ExtractedFile{
		Name:   flatbufferName,
		Size:   uint64(len(jsonData)),
		Reader: bytes.NewReader(jsonData),
	}, nil
}

func extractTableDataCommon(
	ctx context.Context,
	provider ExcelProvider,
	rows *sql.Rows,
	flatbufferName string,
) ([]any, error) {
	var schemaData []any

	for rows.Next() {
		var row []byte
		if err := rows.Scan(&row); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		t := provider.GetExcelByName(strings.ToLower(flatbufferName))
		if t == nil {
			slog.WarnContext(ctx, "Table not found", "name", flatbufferName)
			continue
		}

		if err := t.Unmarshal(row); err != nil {
			return nil, fmt.Errorf("failed to unmarshal row: %w", err)
		}

		schemaData = append(schemaData, t)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to get tables: %w", rows.Err())
	}

	return schemaData, nil
}
