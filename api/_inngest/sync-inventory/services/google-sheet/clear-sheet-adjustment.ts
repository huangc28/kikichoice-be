import { getGoogleSheetsClient } from "./client.js";
import { env } from "#shared/env.js";

export const clearSheetAdjustment = async (processedSkus: string[]) => {
  if (processedSkus.length === 0) {
    console.info("No product adjustment stock to clear");
    return;
  }

  try {
    const client = getGoogleSheetsClient();
    const sheetData = await client.spreadsheets.values.get({
      spreadsheetId: env.GOOGLE_SHEET_ID,
      range: env.GOOGLE_SHEET_RANGE,
    });

    const rows = sheetData.data.values || [];
    const updates = findStockAdjustmentRowsToClear(rows, processedSkus);

    if (updates.length > 0) {
      await client.spreadsheets.values.batchUpdate({
        spreadsheetId: env.GOOGLE_SHEET_ID,
        requestBody: {
          valueInputOption: "RAW",
          data: updates,
        },
      });
    }
    console.log(
      `✅ Successfully cleared adjustment values for SKUs: ${
        processedSkus.join(", ")
      }`,
    );
  } catch (error) {
    console.error("Error clearing stock adjustment:", error);
    throw error;
  }
};

const findStockAdjustmentRowsToClear = (
  rows: any[][],
  skus: string[],
): any[] => {
  const processedSkuSet = new Set(skus);
  const updates: any[] = [];

  rows.forEach((row, index) => {
    const sheetSku = (row[0] || "").trim();
    if (processedSkuSet.has(sheetSku)) {
      const rowNumber = index + 2;
      const adjustmentColumnRange = `G${rowNumber}`;

      updates.push({
        range: adjustmentColumnRange,
        values: [["0"]],
      });
    }
  });

  return updates;
};
