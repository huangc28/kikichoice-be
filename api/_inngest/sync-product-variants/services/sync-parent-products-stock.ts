import { env } from "#shared/env.js";
import { getGoogleSheetsClient } from "#shared/client.js";
import type { ProcessedProductVariant } from "./upsert-product-variants.js";

export const syncParentProductsStock = async (
  processedVariants: ProcessedProductVariant[],
) => {
  if (processedVariants.length === 0) {
    console.info("No processed variants to calculate parent stock");
    return;
  }

  console.log(
    `🔄 Starting parent products stock sync for ${processedVariants.length} variants...`,
  );

  const client = getGoogleSheetsClient();

  // Fetch current parent products sheet data
  const sheetData = await client.spreadsheets.values.get({
    spreadsheetId: env.GOOGLE_SHEET_ID,
    range: env.GOOGLE_SHEET_RANGE,
  });

  const rows = sheetData.data.values || [];
  console.info(`📊 Found ${rows.length} parent products in sheet`);

  // Calculate parent stock totals from variants
  const parentStockTotals = calculateParentStockTotals(processedVariants);
  console.info(
    `📈 Calculated stock for ${parentStockTotals.size} parent products`,
  );

  // Prepare sheet updates
  const sheetName = env.GOOGLE_SHEET_RANGE.split("!")[0];
  const updates = prepareParentStockUpdates(
    sheetName,
    rows,
    parentStockTotals,
  );

  if (updates.length > 0) {
    await client.spreadsheets.values.batchUpdate({
      spreadsheetId: env.GOOGLE_SHEET_ID,
      requestBody: {
        valueInputOption: "RAW",
        data: updates,
      },
    });

    console.log(
      `✅ Successfully updated stock for ${updates.length} parent products in sheet`,
    );
  } else {
    console.log("📭 No matching parent product SKUs found in sheet to update");
  }
};

const calculateParentStockTotals = (
  processedVariants: ProcessedProductVariant[],
): Map<string, number> => {
  const parentStockMap = new Map<string, number>();

  processedVariants.forEach((variant) => {
    if (variant.parent_sku) {
      const currentTotal = parentStockMap.get(variant.parent_sku) || 0;
      parentStockMap.set(
        variant.parent_sku,
        currentTotal + variant.stock_count,
      );
    }
  });

  return parentStockMap;
};

const prepareParentStockUpdates = (
  sheetName: string,
  rows: any[][],
  parentStockTotals: Map<string, number>,
): any[] => {
  const updates: any[] = [];

  rows.forEach((row, index) => {
    const parentSku = (row[0] || "").trim(); // Column A contains the parent SKU
    const calculatedStock = parentStockTotals.get(parentSku);

    if (calculatedStock !== undefined) {
      const rowNumber = index + 2; // +2 for header and 1-based indexing

      // Update stock count in Column H
      updates.push({
        range: `${sheetName}!H${rowNumber}`,
        values: [[calculatedStock.toString()]],
      });

      console.log(
        `📦 Parent ${parentSku}: stock updated to ${calculatedStock}`,
      );
    }
  });

  return updates;
};
