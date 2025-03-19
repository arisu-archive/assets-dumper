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

	"github.com/arisu-archive/arona-flatbuffers/go/excel"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver.
)

func sqliteExtractor(ctx context.Context, inputPath string) (*Result, error) {
	db, err := openDatabase(inputPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tableNames, err := getTableNames(db)
	if err != nil {
		return nil, err
	}

	files, err := processAllTables(ctx, db, tableNames)
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

func openDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}
	return db, nil
}

func getTableNames(db *sql.DB) ([]string, error) {
	tblRows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
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

	return tableNames, nil
}

func processAllTables(ctx context.Context, db *sql.DB, tableNames []string) ([]ExtractedFile, error) {
	files := make([]ExtractedFile, 0, len(tableNames))

	for _, name := range tableNames {
		file, err := processTable(ctx, db, name)
		if err != nil {
			return nil, err
		}

		if file != nil {
			files = append(files, *file)
		}
	}

	return files, nil
}

func processTable(ctx context.Context, db *sql.DB, tableName string) (*ExtractedFile, error) {
	slog.DebugContext(ctx, "Table", "name", strings.TrimSuffix(tableName, "DBSchema"))

	// Some tables contains underscore, and our mapping is from the actual struct name.
	flatbufferName := strings.ReplaceAll(strings.TrimSuffix(tableName, "DBSchema")+"Excel", "_", "")

	// Query the table
	rows, err := db.Query(fmt.Sprintf("SELECT Bytes FROM %s", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}
	defer rows.Close()

	schemaData, err := extractTableData(ctx, rows, flatbufferName)
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

func extractTableData(ctx context.Context, rows *sql.Rows, flatbufferName string) ([]any, error) {
	var schemaData []any

	for rows.Next() {
		var row []byte
		if err := rows.Scan(&row); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		t := excel.GetFlatDataByName(strings.ToLower(flatbufferName))
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
