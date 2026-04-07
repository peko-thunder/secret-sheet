import * as fs from "fs";
import * as path from "path";
import * as dotenv from "dotenv";

/** Sentinel value in .env that marks a variable to be fetched from the spreadsheet */
export const SHEET_PLACEHOLDER = "$sheet";

export interface ParsedEnv {
  /** Variables to fetch from the spreadsheet */
  sheetKeys: string[];
  /** Variables with literal values from .env */
  localVars: Record<string, string>;
}

/**
 * Parse a .env file and separate:
 *   - Keys whose value is exactly "$sheet" → fetch from spreadsheet
 *   - Everything else                       → use the literal value
 *
 * .env example:
 *   DATABASE_URL=postgres://localhost/mydb   # local value
 *   SECRET_API_KEY=$sheet                    # fetched from spreadsheet
 *   AWS_SECRET_ACCESS_KEY=$sheet             # fetched from spreadsheet
 */
export function parseEnvFile(envFilePath: string): ParsedEnv {
  const fullPath = path.resolve(process.cwd(), envFilePath);

  if (!fs.existsSync(fullPath)) {
    return { sheetKeys: [], localVars: {} };
  }

  const content = fs.readFileSync(fullPath, "utf-8");
  const parsed = dotenv.parse(content);

  const sheetKeys: string[] = [];
  const localVars: Record<string, string> = {};

  for (const [key, value] of Object.entries(parsed)) {
    if (value === SHEET_PLACEHOLDER) {
      sheetKeys.push(key);
    } else {
      localVars[key] = value;
    }
  }

  return { sheetKeys, localVars };
}
