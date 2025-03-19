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

	// sqlite3 driver.
	"github.com/arisu-archive/arona-flatbuffers/go/excel"
	_ "github.com/mattn/go-sqlite3"
)

func sqliteExtractor(ctx context.Context, inputPath string) (*Result, error) {
	db, err := sql.Open("sqlite3", inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}
	defer db.Close()

	// Printout all the tables in the database
	tblRows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer tblRows.Close()

	files := make([]ExtractedFile, 0)
	// As we are not planning to query the data. We can simply wrap it with DataList and convert it into JSON.
	for tblRows.Next() {
		var name string
		if scanErr := tblRows.Scan(&name); scanErr != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", scanErr)
		}
		slog.DebugContext(ctx, "Table", "name", strings.TrimSuffix(name, "DBSchema"))
		// Some tables contains underscore, and our mapping is from the actual struct name.
		flatbufferName := strings.ReplaceAll(strings.TrimSuffix(name, "DBSchema")+"Excel", "_", "")

		// Query the table
		rows, err := db.Query(fmt.Sprintf("SELECT Bytes FROM %s", name))
		if err != nil {
			return nil, fmt.Errorf("failed to query table: %w", err)
		}
		defer rows.Close()

		// Read the table
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
			if err != nil {
				return nil, fmt.Errorf("failed to marshal row: %w", err)
			}
			schemaData = append(schemaData, t)
		}
		if rows.Err() != nil {
			return nil, fmt.Errorf("failed to get tables: %w", rows.Err())
		}
		jsonData, err := json.MarshalIndent(map[string]any{
			"data_list": schemaData,
		}, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal row: %w", err)
		}
		files = append(files, ExtractedFile{
			Name:   flatbufferName,
			Size:   len(jsonData),
			Reader: bytes.NewReader(jsonData),
		})
	}
	if tblRows.Err() != nil {
		return nil, fmt.Errorf("failed to get tables: %w", tblRows.Err())
	}

	fileName := filepath.Base(inputPath)
	folder := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return &Result{
		rc:     db,
		Folder: folder,
		Files:  files,
	}, nil
}
