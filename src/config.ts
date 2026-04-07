import * as fs from "fs";
import * as path from "path";
import * as dotenv from "dotenv";

export interface SheetEnvConfig {
  spreadsheetId: string;
  sheetName: string;
  keyColumn: string;
  valueColumn: string;
  credentialsFile?: string;
  headerRow: boolean;
}

/**
 * Load config from .sheetenv file and/or environment variables.
 *
 * Priority (high → low):
 *   1. Environment variables already set in the shell
 *   2. .sheetenv config file in the current working directory
 */
export function loadConfig(configPath: string): SheetEnvConfig {
  const fullPath = path.resolve(process.cwd(), configPath);

  // Load config file values into process.env (only if not already set)
  if (fs.existsSync(fullPath)) {
    dotenv.config({ path: fullPath, override: false });
  }

  const spreadsheetId = process.env.SHEET_ENV_SPREADSHEET_ID;
  if (!spreadsheetId) {
    throw new Error(
      [
        "SHEET_ENV_SPREADSHEET_ID is not set.",
        `Create a "${configPath}" file in your project root with:`,
        "",
        "  SHEET_ENV_SPREADSHEET_ID=your-spreadsheet-id",
        "  SHEET_ENV_SHEET_NAME=Sheet1",
        "  SHEET_ENV_CREDENTIALS=/path/to/credentials.json",
        "",
        "Or export SHEET_ENV_SPREADSHEET_ID in your shell.",
      ].join("\n")
    );
  }

  return {
    spreadsheetId,
    sheetName: process.env.SHEET_ENV_SHEET_NAME ?? "Sheet1",
    keyColumn: process.env.SHEET_ENV_KEY_COLUMN ?? "A",
    valueColumn: process.env.SHEET_ENV_VALUE_COLUMN ?? "B",
    credentialsFile:
      process.env.SHEET_ENV_CREDENTIALS ??
      process.env.GOOGLE_APPLICATION_CREDENTIALS,
    headerRow: (process.env.SHEET_ENV_HEADER_ROW ?? "false") === "true",
  };
}
