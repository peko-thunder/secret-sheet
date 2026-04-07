package main

import (
	"os"
	"path/filepath"
	"strings"
)

// sheetPlaceholder is the sentinel value in .env that means
// "fetch this variable from the spreadsheet".
const sheetPlaceholder = "$sheet"

// parseEnvFile reads a .env file and splits its contents into:
//   - sheetKeys : variable names whose value is exactly "$sheet"
//   - localVars : all other key=value pairs
//
// Example .env:
//
//	DATABASE_URL=postgres://localhost/mydb   # used as-is
//	SECRET_API_KEY=$sheet                    # fetched from spreadsheet
func parseEnvFile(envPath string) (sheetKeys []string, localVars map[string]string, err error) {
	abs, err := filepath.Abs(envPath)
	if err != nil {
		return nil, nil, err
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, map[string]string{}, nil
		}
		return nil, nil, err
	}

	localVars = make(map[string]string)

	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		// Inline comment after unquoted value (e.g. KEY=value # comment)
		if val != "" && val[0] != '"' && val[0] != '\'' {
			if ci := strings.Index(val, " #"); ci >= 0 {
				val = strings.TrimSpace(val[:ci])
			}
		} else {
			val = stripQuotes(val)
		}

		if val == sheetPlaceholder {
			sheetKeys = append(sheetKeys, key)
		} else {
			localVars[key] = val
		}
	}

	return sheetKeys, localVars, nil
}
