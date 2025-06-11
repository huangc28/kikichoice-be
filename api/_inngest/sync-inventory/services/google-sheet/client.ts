import { google } from "googleapis";
import { env } from "#shared/env.js";

export const getGoogleSheetsClient = () => {
  const auth = new google.auth.GoogleAuth({
    credentials: {
      client_email: env.GOOGLE_SERVICE_ACCOUNT_EMAIL,
      private_key: env.GOOGLE_SERVICE_ACCOUNT_PRIVATE_KEY?.replace(
        /\\n/g,
        "\n",
      ),
    },
    scopes: ["https://www.googleapis.com/auth/spreadsheets"],
  });

  return google.sheets({ version: "v4", auth });
};
