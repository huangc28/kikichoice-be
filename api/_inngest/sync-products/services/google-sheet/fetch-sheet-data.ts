import { env } from "#shared/env.js";
import { getGoogleSheetsClient } from "#shared/client.js";

import type { ProductRow } from "../../types.js";

// Function to fetch parent SKUs from variants sheet
const fetchParentSkusWithVariants = async (): Promise<Set<string>> => {
  try {
    const sheets = getGoogleSheetsClient();
    const spreadsheetId = env.GOOGLE_PROD_VARIANTS_SHEET_ID;
    const range = env.GOOGLE_PROD_VARIANTS_SHEET_RANGE || "Sheet1!A2:A1000";

    const response = await sheets.spreadsheets.values.get({
      spreadsheetId,
      range,
    });

    const rows = response.data.values || [];

    // Extract parent SKUs from column A (母商品 sku)
    const parentSkus = new Set<string>();
    rows.forEach((row) => {
      const parentSku = (row[0] || "").trim();
      if (parentSku) {
        parentSkus.add(parentSku);
      }
    });

    console.info(`Found ${parentSkus.size} parent products with variants`);
    return parentSkus;
  } catch (error) {
    console.error("Error fetching variants data:", error);
    // Return empty set to continue with normal processing if variants fetch fails
    console.warn("Continuing without variant filtering due to fetch error");
    return new Set<string>();
  }
};

// Function to fetch data from Google Sheets
export const fetchSheetData = async (): Promise<ProductRow[]> => {
  try {
    const sheets = getGoogleSheetsClient();
    const spreadsheetId = env.GOOGLE_SHEET_ID;
    const range = env.GOOGLE_SHEET_RANGE || "Sheet1!A2:H1000"; // Skip header row

    // Fetch both products and variants data in parallel
    const [productsResponse, parentSkusWithVariants] = await Promise.all([
      sheets.spreadsheets.values.get({
        spreadsheetId,
        range,
      }),
      fetchParentSkusWithVariants(),
    ]);

    const rows = productsResponse.data.values || [];

    console.info("number of products", rows.length);

    return rows
      .map((row) => transformRowToProduct(row, parentSkusWithVariants));
  } catch (error) {
    console.error("Error fetching sheet data:", error);
    throw new Error(`Failed to fetch sheet data: ${error}`);
  }
};

function transformRowToProduct(
  row: string[],
  parentSkusWithVariants: Set<string>,
): ProductRow {
  const sku = (row[0] || "").trim();
  const stockAdjustCount = parseInt((row[6] || "0").trim()) || 0;

  // If this product has variants, skip stock sync by setting stock_adjust_count to 0
  const finalStockAdjustCount = parentSkusWithVariants.has(sku)
    ? 0
    : stockAdjustCount;

  if (parentSkusWithVariants.has(sku) && stockAdjustCount !== 0) {
    console.info(
      `⚠️ Skipping stock sync for product ${sku} (has variants, original adjustment: ${stockAdjustCount})`,
    );
  }

  return {
    sku,
    name: (row[1] || "").trim(),
    ready_for_sale: (row[2] || "").trim() === "Y",
    short_desc: (row[3] || "").trim(),
    stock_adjust_count: finalStockAdjustCount,
    price: parseFloat((row[11] || "0").trim()) || 0,
  };
}
