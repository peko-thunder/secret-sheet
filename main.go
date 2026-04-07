package main

import (
	"fmt"
	"os"
)

var version = "dev" // overridden by -ldflags at build time

func main() {
	opts, cmd, err := parseArgs(os.Args[1:])
	if err != nil {
		fatalf("%v", err)
	}

	if len(cmd) == 0 {
		printHelp()
		os.Exit(1)
	}

	// ── 1. Load tool config (.sheetenv) ──────────────────────────────────────
	cfg, err := loadConfig(opts.configFile)
	if err != nil {
		fatalf("%v", err)
	}
	logf(opts.verbose, "Spreadsheet: %s  Sheet: %s", cfg.SpreadsheetID, cfg.SheetName)

	// ── 2. Parse project .env ────────────────────────────────────────────────
	sheetKeys, localVars, err := parseEnvFile(opts.envFile)
	if err != nil {
		fatalf("reading %s: %v", opts.envFile, err)
	}
	logf(opts.verbose, "Local vars: %d  Sheet keys: %d", len(localVars), len(sheetKeys))

	// ── 3. Fetch secrets from spreadsheet ────────────────────────────────────
	fetchedSecrets := map[string]string{}
	if len(sheetKeys) > 0 {
		fmt.Fprintf(os.Stderr, "[sheet-env] Fetching %d secret(s) from spreadsheet...\n", len(sheetKeys))

		fetchedSecrets, err = fetchSecrets(cfg, sheetKeys)
		if err != nil {
			fatalf("fetching secrets: %v", err)
		}

		for _, key := range sheetKeys {
			if _, ok := fetchedSecrets[key]; !ok {
				fmt.Fprintf(os.Stderr, "[sheet-env] Warning: %q not found in spreadsheet\n", key)
			} else {
				logf(opts.verbose, "Fetched: %s", key)
			}
		}
	}

	// ── 4. Merge env vars (fetched secrets take precedence) ──────────────────
	merged := make(map[string]string, len(localVars)+len(fetchedSecrets))
	for k, v := range localVars {
		merged[k] = v
	}
	for k, v := range fetchedSecrets {
		merged[k] = v
	}

	// ── 5. Run wrapped command ────────────────────────────────────────────────
	logf(opts.verbose, "Running: %v", cmd)

	if err := runCommand(cmd[0], cmd[1:], merged); err != nil {
		fatalf("%v", err)
	}
}

// ─── CLI options ──────────────────────────────────────────────────────────────

type cliOpts struct {
	envFile    string
	configFile string
	verbose    bool
}

func parseArgs(argv []string) (cliOpts, []string, error) {
	opts := cliOpts{
		envFile:    ".env",
		configFile: ".sheetenv",
	}

	i := 0
	for i < len(argv) {
		arg := argv[i]

		switch {
		case arg == "--":
			return opts, argv[i+1:], nil

		case arg == "-h", arg == "--help":
			printHelp()
			os.Exit(0)

		case arg == "-v", arg == "--version":
			fmt.Println(version)
			os.Exit(0)

		case arg == "--verbose":
			opts.verbose = true

		case arg == "-e", arg == "--env-file":
			if i+1 >= len(argv) {
				return opts, nil, fmt.Errorf("%s requires a value", arg)
			}
			i++
			opts.envFile = argv[i]

		case len(arg) > 11 && arg[:11] == "--env-file=":
			opts.envFile = arg[11:]

		case arg == "-c", arg == "--config":
			if i+1 >= len(argv) {
				return opts, nil, fmt.Errorf("%s requires a value", arg)
			}
			i++
			opts.configFile = argv[i]

		case len(arg) > 9 && arg[:9] == "--config=":
			opts.configFile = arg[9:]

		default:
			// First non-option token starts the wrapped command
			return opts, argv[i:], nil
		}
		i++
	}

	return opts, nil, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[sheet-env] Error: "+format+"\n", args...)
	os.Exit(1)
}

func logf(verbose bool, format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stderr, "[sheet-env] "+format+"\n", args...)
	}
}

func printHelp() {
	fmt.Print(`
sheet-env — inject secrets from Google Sheets as environment variables

Usage:
  sheet-env [options] [--] <command> [args...]

Options:
  -e, --env-file <path>   .env file to read  (default: .env)
  -c, --config <path>     Config file        (default: .sheetenv)
      --verbose           Print debug info to stderr
  -h, --help              Show this help
  -v, --version           Print version

Examples:
  sheet-env npm run dev
  sheet-env python app.py
  sheet-env -- node server.js --port 3000
  sheet-env -e .env.production -- npm start

─────────────────────────────────────────────────
.env file  (in your project root)
─────────────────────────────────────────────────
  DATABASE_URL=postgres://localhost/mydb   # plain value — used as-is
  SECRET_API_KEY=$sheet                    # fetched from spreadsheet
  AWS_SECRET_ACCESS_KEY=$sheet             # fetched from spreadsheet

─────────────────────────────────────────────────
.sheetenv config file  (in your project root)
─────────────────────────────────────────────────
  SHEET_ENV_SPREADSHEET_ID=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
  SHEET_ENV_SHEET_NAME=Sheet1
  SHEET_ENV_CREDENTIALS=/path/to/credentials.json  # optional

─────────────────────────────────────────────────
Authentication
─────────────────────────────────────────────────
  Option A — Service account (teams / CI):
    Set SHEET_ENV_CREDENTIALS to the path of a service account JSON key.

  Option B — Application Default Credentials (personal):
    Run: gcloud auth application-default login
    No SHEET_ENV_CREDENTIALS needed.
`)
}
