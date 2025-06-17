import { env } from "#shared/env.js";
import { getGoogleSheetsClient } from "#shared/client.js";
import { ProductVariant } from "../types.js";

export const fetchProductVariantsFromSheet = async () => {
  const sheets = getGoogleSheetsClient();
  const spreadsheetId = env.GOOGLE_PROD_VARIANTS_SHEET_ID;
  const range = env.GOOGLE_PROD_VARIANTS_SHEET_RANGE || "Sheet1!A2:H1000"; // Skip header row

  const response = await sheets.spreadsheets.values.get({
    spreadsheetId,
    range,
  });

  const rows = response.data.values || [];
  console.info("number of product variants", rows.length);

  return rows.map(transformRowToProductVariant);
};

function transformRowToProductVariant(row: string[]): ProductVariant {
  return {
    parent_sku: (row[0] || "").trim(),
    sku: (row[1] || "").trim(),
    name: (row[2] || "").trim(),
    stock_adjust_count: parseInt((row[3] || "0").trim()) || 0,
    price: parseFloat((row[5] || "0").trim()) || 0,
  };
}
