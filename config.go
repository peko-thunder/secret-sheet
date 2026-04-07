package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all runtime settings for sheet-env.
type Config struct {
	SpreadsheetID   string
	SheetName       string
	KeyColumn       string
	ValueColumn     string
	CredentialsFile string
	HeaderRow       bool
}

// loadConfig reads the .sheetenv config file (if present) and merges it with
// the process environment. Shell variables always take precedence over the file.
func loadConfig(configPath string) (*Config, error) {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	if data, err := os.ReadFile(abs); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			idx := strings.IndexByte(line, '=')
			if idx < 0 {
				continue
			}
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			val = stripQuotes(val)
			// Only set if not already present in the environment
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, val) //nolint:errcheck
			}
		}
	}

	id := os.Getenv("SHEET_ENV_SPREADSHEET_ID")
	if id == "" {
		return nil, fmt.Errorf(
			"SHEET_ENV_SPREADSHEET_ID is not set.\n"+
				"Create a %q file in your project root:\n\n"+
				"  SHEET_ENV_SPREADSHEET_ID=your-spreadsheet-id\n"+
				"  SHEET_ENV_SHEET_NAME=Sheet1\n"+
				"  SHEET_ENV_CREDENTIALS=/path/to/credentials.json\n",
			configPath,
		)
	}

	return &Config{
		SpreadsheetID:   id,
		SheetName:       envOr("SHEET_ENV_SHEET_NAME", "Sheet1"),
		KeyColumn:       envOr("SHEET_ENV_KEY_COLUMN", "A"),
		ValueColumn:     envOr("SHEET_ENV_VALUE_COLUMN", "B"),
		CredentialsFile: envOr("SHEET_ENV_CREDENTIALS", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")),
		HeaderRow:       os.Getenv("SHEET_ENV_HEADER_ROW") == "true",
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
