import { getGoogleSheetsClient } from "./client.js";
import { env } from "#shared/env.js";

export type StockUpdate = {
  sku: string;
  stock: number;
};

export const syncSheetCurrentStock = async (stockUpdates: StockUpdate[]) => {
  if (stockUpdates.length === 0) {
    console.info("No stock updates to sync");
    return;
  }

  const client = getGoogleSheetsClient();
  const sheetData = await client.spreadsheets.values.get({
    spreadsheetId: env.GOOGLE_SHEET_ID,
    range: env.GOOGLE_SHEET_RANGE,
  });

  const rows = sheetData.data.values || [];
  const updates = prepareSheetUpdates(rows, stockUpdates);

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
    `✅ Successfully synced current stock for ${stockUpdates.length} products`,
  );
};

const prepareSheetUpdates = (
  rows: any[][],
  stockUpdates: StockUpdate[],
): any[] => {
  const stockUpdateMap = new Map(
    stockUpdates.map((update) => [update.sku, update.stock]),
  );

  const updates: any[] = [];

  rows.forEach((row, index) => {
    const sheetSku = (row[0] || "").trim();
    const newStock = stockUpdateMap.get(sheetSku);

    if (newStock !== undefined) {
      const rowNumber = index + 2; // +2 for header and 1-based indexing

      // Update 目前庫存 (Column H)
      updates.push({
        range: `H${rowNumber}`,
        values: [[newStock.toString()]],
      });

      // Clear adjustment count (Column G)
      updates.push({
        range: `G${rowNumber}`,
        values: [["0"]],
      });
    }
  });

  return updates;
};
