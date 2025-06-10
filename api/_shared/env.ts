// Import zod with ES module syntax for better TypeScript support
import { z } from "zod";

/**
 * Environment variable schema with validation and default values
 */
const envSchema = z.object({
  VERCEL_ENV: z
    .enum(["development", "production", "preview"])
    .default("development"),

  // Database configuration
  DB_HOST: z.string().default(""),
  DB_USER: z.string().default(""),
  DB_PASSWORD: z.string().default(""),
  DB_NAME: z.string().default(""),
  DB_PORT: z
    .string()
    .transform((val: string) => Number(val) || 5432)
    .default("55322"),

  AZURE_BLOB_STORAGE_ACCOUNT_NAME: z.string().default(""),
  AZURE_BLOB_STORAGE_KEY: z.string().default(""),
  AZURE_BLOB_STORAGE_CONNECTION_STRING: z.string().default(""),

  TELEGRAM_BOT_TOKEN: z.string().default(""),

  // Google Sheets Configuration
  GOOGLE_SHEET_ID: z.string().default(""),
  GOOGLE_SHEET_RANGE: z.string().default(""),
  GOOGLE_SERVICE_ACCOUNT_EMAIL: z.string().default(""),
  GOOGLE_SERVICE_ACCOUNT_PRIVATE_KEY: z.string().default(""),

  INNGEST_APP_ID: z.string().default(""),
});

// Infer the TypeScript type from the Zod schema
export type Env = z.infer<typeof envSchema>;

/**
 * Load and validate environment variables
 *
 * @returns Validated environment object or undefined if validation fails
 */
export function getEnv(): Env {
  const result = envSchema.safeParse(process.env);

  const envData = result.data!;
  const missingVars: string[] = [];

  // Check each environment variable for empty strings
  Object.entries(envData).forEach(([key, value]) => {
    if (typeof value === "string" && value.trim() === "") {
      missingVars.push(key);
    }
  });

  if (missingVars.length > 0) {
    console.error("❌ Missing environment variables (empty values detected):");
    missingVars.forEach((varName) => {
      console.error(`  • ${varName}: is empty (missing)`);
    });
    console.error("⚠️  Application will continue with default empty values");
  }

  return envData;
}

export const env = getEnv();
