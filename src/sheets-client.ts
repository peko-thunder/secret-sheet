import * as fs from "fs";
import { google } from "googleapis";
import type { SheetEnvConfig } from "./config";

/**
 * Fetch the specified keys from a Google Spreadsheet.
 *
 * Expected spreadsheet layout:
 *   Column A (keyColumn)   : variable name  (e.g. "SECRET_API_KEY")
 *   Column B (valueColumn) : secret value   (e.g. "sk-abc123...")
 *
 * Rows whose key column is empty are skipped.
 * If headerRow is true in config, the first row is skipped.
 */
export async function fetchSecrets(
  config: SheetEnvConfig,
  keys: string[]
): Promise<Record<string, string>> {
  const auth = await buildAuth(config);
  const sheets = google.sheets({ version: "v4", auth });

  const range = `${config.sheetName}!${config.keyColumn}:${config.valueColumn}`;

  const response = await sheets.spreadsheets.values.get({
    spreadsheetId: config.spreadsheetId,
    range,
  });

  const rows = response.data.values ?? [];
  const result: Record<string, string> = {};

  const startIndex = config.headerRow ? 1 : 0;

  for (let i = startIndex; i < rows.length; i++) {
    const row = rows[i];
    const key = row[0];
    const value = row[1];

    if (!key || value === undefined || value === null) continue;
    if (keys.includes(key)) {
      result[key] = String(value);
    }
  }

  return result;
}

async function buildAuth(config: SheetEnvConfig) {
  const scopes = ["https://www.googleapis.com/auth/spreadsheets.readonly"];

  if (config.credentialsFile) {
    const raw = fs.readFileSync(config.credentialsFile, "utf-8");
    const credentials = JSON.parse(raw) as Record<string, unknown>;

    // Service account key file contains "type": "service_account"
    return new google.auth.GoogleAuth({ credentials, scopes });
  }

  // Fall back to Application Default Credentials
  // (works with: gcloud auth application-default login)
  return new google.auth.GoogleAuth({ scopes });
}
