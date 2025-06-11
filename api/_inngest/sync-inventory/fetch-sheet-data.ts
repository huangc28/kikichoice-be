import { google } from "googleapis";
import { env } from "#shared/env.js";

export interface ProductRow {
  sku: string;
  name: string;
  ready_for_sale: boolean;
  stock_count: number;
  price: number;
  short_desc: string;
}

// Google Sheets API client setup
const getGoogleSheetsClient = () => {
  const auth = new google.auth.GoogleAuth({
    credentials: {
      client_email: env.GOOGLE_SERVICE_ACCOUNT_EMAIL,
      private_key: env.GOOGLE_SERVICE_ACCOUNT_PRIVATE_KEY?.replace(
        /\\n/g,
        "\n",
      ),
    },
    scopes: ["https://www.googleapis.com/auth/spreadsheets.readonly"],
  });

  return google.sheets({ version: "v4", auth });
};

// Function to fetch data from Google Sheets
export const fetchSheetData = async (): Promise<ProductRow[]> => {
  try {
    const sheets = getGoogleSheetsClient();
    const spreadsheetId = process.env.GOOGLE_SHEET_ID;
    const range = process.env.GOOGLE_SHEET_RANGE || "Sheet1!A2:H1000"; // Skip header row

    console.info("** 1", process.env.GOOGLE_SHEET_RANGE);

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
    stock_count: parseInt((row[6] || "0").trim()) || 0,
    price: parseFloat((row[10] || "0").trim()) || 0,
  };
}
