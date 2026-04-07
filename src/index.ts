#!/usr/bin/env node
import { loadConfig } from "./config";
import { parseEnvFile } from "./env-parser";
import { fetchSecrets } from "./sheets-client";
import { runCommand } from "./runner";

// ──────────────────────────────────────────────
// Argument parsing
// ──────────────────────────────────────────────

interface CliOptions {
  envFile: string;
  configFile: string;
  verbose: boolean;
}

function parseArgs(argv: string[]): { options: CliOptions; command: string[] } {
  const options: CliOptions = {
    envFile: ".env",
    configFile: ".sheetenv",
    verbose: false,
  };

  let i = 0;

  while (i < argv.length) {
    const arg = argv[i];

    if (arg === "--") {
      i++;
      break;
    }

    if (arg === "--help" || arg === "-h") {
      printHelp();
      process.exit(0);
    }

    if (arg === "--version" || arg === "-v") {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const pkg = require("../package.json") as { version: string };
      console.log(pkg.version);
      process.exit(0);
    }

    if (arg === "--verbose") {
      options.verbose = true;
      i++;
      continue;
    }

    if (arg === "--env-file" || arg === "-e") {
      if (!argv[i + 1]) fatal("--env-file requires a value");
      options.envFile = argv[i + 1];
      i += 2;
      continue;
    }

    if (arg.startsWith("--env-file=")) {
      options.envFile = arg.slice("--env-file=".length);
      i++;
      continue;
    }

    if (arg === "--config" || arg === "-c") {
      if (!argv[i + 1]) fatal("--config requires a value");
      options.configFile = argv[i + 1];
      i += 2;
      continue;
    }

    if (arg.startsWith("--config=")) {
      options.configFile = arg.slice("--config=".length);
      i++;
      continue;
    }

    // First non-option argument → start of the wrapped command
    break;
  }

  const command = argv.slice(i);
  return { options, command };
}

// ──────────────────────────────────────────────
// Main
// ──────────────────────────────────────────────

async function main() {
  const { options, command } = parseArgs(process.argv.slice(2));

  if (command.length === 0) {
    printHelp();
    process.exit(1);
  }

  const log = options.verbose
    ? (msg: string) => process.stderr.write(`[sheet-env] ${msg}\n`)
    : (_msg: string) => {};

  // 1. Load tool config (.sheetenv)
  const config = loadConfig(options.configFile);
  log(`Spreadsheet: ${config.spreadsheetId}  Sheet: ${config.sheetName}`);

  // 2. Parse project .env
  const { sheetKeys, localVars } = parseEnvFile(options.envFile);
  log(`Local vars: ${Object.keys(localVars).length}  Sheet keys: ${sheetKeys.length}`);

  // 3. Fetch secrets from spreadsheet
  let fetchedSecrets: Record<string, string> = {};

  if (sheetKeys.length > 0) {
    process.stderr.write(
      `[sheet-env] Fetching ${sheetKeys.length} secret(s) from spreadsheet...\n`
    );

    fetchedSecrets = await fetchSecrets(config, sheetKeys);

    // Warn about any keys not found in the sheet
    for (const key of sheetKeys) {
      if (!(key in fetchedSecrets)) {
        process.stderr.write(
          `[sheet-env] Warning: "${key}" was not found in the spreadsheet\n`
        );
      } else {
        log(`Fetched: ${key}`);
      }
    }
  }

  // 4. Build final env: local .env values + fetched secrets
  //    Fetched secrets take precedence over local values for the same key.
  const envVars: Record<string, string> = {
    ...localVars,
    ...fetchedSecrets,
  };

  // 5. Run the wrapped command
  const [cmd, ...args] = command;
  log(`Running: ${cmd} ${args.join(" ")}`);

  await runCommand(cmd, args, envVars);
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

function fatal(msg: string): never {
  process.stderr.write(`[sheet-env] Error: ${msg}\n`);
  process.exit(1);
}

function printHelp() {
  console.log(`
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

Variables whose value is exactly "$sheet" are fetched from the spreadsheet.
All other variables are passed through as normal env vars.

─────────────────────────────────────────────────
.sheetenv config file  (in your project root)
─────────────────────────────────────────────────
  SHEET_ENV_SPREADSHEET_ID=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
  SHEET_ENV_SHEET_NAME=Sheet1
  SHEET_ENV_CREDENTIALS=/path/to/credentials.json

  # Optional
  SHEET_ENV_KEY_COLUMN=A          # column for variable names (default: A)
  SHEET_ENV_VALUE_COLUMN=B        # column for values         (default: B)
  SHEET_ENV_HEADER_ROW=false      # skip first row as header  (default: false)

─────────────────────────────────────────────────
Spreadsheet layout
─────────────────────────────────────────────────
  A                      B
  ─────────────────────────────────
  SECRET_API_KEY         sk-abc123...
  AWS_SECRET_ACCESS_KEY  AKIA...
  DATABASE_PASSWORD      hunter2

─────────────────────────────────────────────────
Authentication
─────────────────────────────────────────────────
  Option 1 — Service account (recommended for teams):
    Create a service account in Google Cloud Console, share the spreadsheet
    with the service account email, download the JSON key, then set:
      SHEET_ENV_CREDENTIALS=/path/to/service-account.json

  Option 2 — Application Default Credentials (easy for personal use):
    Run: gcloud auth application-default login
    No SHEET_ENV_CREDENTIALS needed.
`);
}

main().catch((err: unknown) => {
  const message = err instanceof Error ? err.message : String(err);
  process.stderr.write(`[sheet-env] Error: ${message}\n`);
  process.exit(1);
});
