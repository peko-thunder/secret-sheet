package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

// fetchSecrets looks up the given keys in the spreadsheet and returns a
// map of key → value for every key that was found.
func fetchSecrets(cfg *Config, keys []string) (map[string]string, error) {
	ctx := context.Background()

	svc, err := buildSheetsService(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// e.g. "Sheet1!A:B"
	rangeStr := fmt.Sprintf("%s!%s:%s", cfg.SheetName, cfg.KeyColumn, cfg.ValueColumn)

	resp, err := svc.Spreadsheets.Values.Get(cfg.SpreadsheetID, rangeStr).Do()
	if err != nil {
		return nil, fmt.Errorf("spreadsheet read failed: %w", err)
	}

	// Build a set for O(1) lookup
	want := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		want[k] = struct{}{}
	}

	startRow := 0
	if cfg.HeaderRow {
		startRow = 1
	}

	result := make(map[string]string)
	for i := startRow; i < len(resp.Values); i++ {
		row := resp.Values[i]
		if len(row) < 2 {
			continue
		}
		key, ok := row[0].(string)
		if !ok || key == "" {
			continue
		}
		if _, needed := want[key]; !needed {
			continue
		}
		switch v := row[1].(type) {
		case string:
			result[key] = v
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}

	return result, nil
}

func buildSheetsService(ctx context.Context, cfg *Config) (*sheets.Service, error) {
	scope := sheets.SpreadsheetsReadonlyScope

	if cfg.CredentialsFile != "" {
		data, err := os.ReadFile(cfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("reading credentials %q: %w", cfg.CredentialsFile, err)
		}
		creds, err := google.CredentialsFromJSON(ctx, data, scope)
		if err != nil {
			return nil, fmt.Errorf("parsing credentials: %w", err)
		}
		return sheets.NewService(ctx, option.WithCredentials(creds))
	}

	// Fall back to Application Default Credentials
	// (works after: gcloud auth application-default login)
	creds, err := google.FindDefaultCredentials(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf(
			"no credentials found: set SHEET_ENV_CREDENTIALS or run "+
				"`gcloud auth application-default login`: %w", err)
	}
	return sheets.NewService(ctx, option.WithCredentials(creds))
}
