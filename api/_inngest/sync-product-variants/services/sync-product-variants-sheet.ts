import { env } from "#shared/env.js";
import { getGoogleSheetsClient } from "#shared/client.js";
import type { ProcessedProductVariant } from "./upsert-product-variants.js";

export const syncProductVariantsSheet = async (
  processedVariants: ProcessedProductVariant[],
) => {
  if (processedVariants.length === 0) {
    console.info("No product variants to sync to sheet");
    return;
  }

  console.log(
    `🔄 Starting sheet sync for ${processedVariants.length} product variants...`,
  );

  const client = getGoogleSheetsClient();
  const sheetData = await client.spreadsheets.values.get({
    spreadsheetId: env.GOOGLE_PROD_VARIANTS_SHEET_ID,
    range: env.GOOGLE_PROD_VARIANTS_SHEET_RANGE,
  });

  const rows = sheetData.data.values || [];
  const sheetName = env.GOOGLE_PROD_VARIANTS_SHEET_RANGE.split("!")[0];
  const updates = prepareSheetUpdates(
    sheetName,
    rows,
    processedVariants,
  );

  if (updates.length > 0) {
    await client.spreadsheets.values.batchUpdate({
      spreadsheetId: env.GOOGLE_PROD_VARIANTS_SHEET_ID,
      requestBody: {
        valueInputOption: "RAW",
        data: updates,
      },
    });

    console.log(
      `✅ Successfully synced ${updates.length / 3} product variants to sheet`,
    );
  } else {
    console.log("📭 No matching SKUs found in sheet to update");
  }
};

const prepareSheetUpdates = (
  sheetName: string,
  rows: any[][],
  processedVariants: ProcessedProductVariant[],
): any[] => {
  const variantUpdateMap = new Map(
    processedVariants.map((variant) => [variant.sku, variant]),
  );

  const updates: any[] = [];

  rows.forEach((row, index) => {
    const sheetSku = (row[1] || "").trim(); // Column B (index 1) contains the SKU
    const variantData = variantUpdateMap.get(sheetSku);

    if (variantData) {
      const rowNumber = index + 2; // +2 for header and 1-based indexing

      // Reset adjust stock (Column D)
      updates.push({
        range: `${sheetName}!D${rowNumber}`,
        values: [["0"]],
      });

      // Update stock_count (Column E)
      updates.push({
        range: `${sheetName}!E${rowNumber}`,
        values: [[variantData.stock_count.toString()]],
      });

      // Update price (Column F)
      updates.push({
        range: `${sheetName}!F${rowNumber}`,
        values: [[variantData.price.toString()]],
      });
    }
  });

  return updates;
};
