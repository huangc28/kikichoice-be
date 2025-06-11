import { env } from "#shared/env.js";
import type { ProductRow } from "../../types.js";
import { getGoogleSheetsClient } from "./client.js";

// Function to fetch data from Google Sheets
export const fetchSheetData = async (): Promise<ProductRow[]> => {
  try {
    const sheets = getGoogleSheetsClient();
    const spreadsheetId = env.GOOGLE_SHEET_ID;
    const range = env.GOOGLE_SHEET_RANGE || "Sheet1!A2:H1000"; // Skip header row

    const response = await sheets.spreadsheets.values.get({
      spreadsheetId,
      range,
    });

    const rows = response.data.values || [];

    console.info("number of products", rows.length);

    return rows
      .map(transformRowToProduct);
  } catch (error) {
    console.error("Error fetching sheet data:", error);
    throw new Error(`Failed to fetch sheet data: ${error}`);
  }
};

function transformRowToProduct(row: string[]): ProductRow {
  return {
    sku: (row[0] || "").trim(),
    name: (row[1] || "").trim(),
    ready_for_sale: (row[2] || "").trim() === "Y",
    short_desc: (row[3] || "").trim(),
    stock_adjust_count: parseInt((row[6] || "0").trim()) || 0,
    price: parseFloat((row[11] || "0").trim()) || 0,
  };
}
