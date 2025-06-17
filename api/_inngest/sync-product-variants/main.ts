import { Inngest, NonRetriableError } from "inngest";
import type { GetFunctionInput } from "inngest";
import { tryCatch } from "#shared/try-catch.js";

import { env } from "#shared/env.js";
import { fetchProductVariantsFromSheet } from "./services/fetch-product-variants.js";
import { syncProductVariants as upsertProductVariants } from "./services/upsert-product-variants.js";
import { syncProductVariantsSheet } from "./services/sync-product-variants-sheet.js";

export const syncProductVariantsFunc = async (
  { step }: GetFunctionInput<Inngest>,
) => {
  const [productVariants, fetchError] = await tryCatch(
    step.run("fetch-sheet-data", fetchProductVariantsFromSheet),
  );

  if (fetchError) {
    console.error("❌ Error fetching product variants:", fetchError);
    throw new NonRetriableError(
      `Error fetching product variants: ${fetchError}`,
    );
  }

  if (!productVariants || productVariants.length === 0) {
    console.info("📭 No product variants to update");
    return;
  }

  console.info(
    `📊 Fetched ${productVariants.length} product variants from sheet`,
  );

  const [result, upsertError] = await tryCatch(
    step.run(
      "upsert-product-variants",
      () => upsertProductVariants(productVariants),
    ),
  );
  if (upsertError) {
    console.error("❌ Error upserting product variants:", upsertError);
    throw upsertError;
  }

  console.info("✅ product variants synced to database");

  // Sync the processed data back to the sheet
  const [, syncSheetError] = await tryCatch(
    step.run(
      "sync-product-variants-sheet",
      () => syncProductVariantsSheet(result!.processedVariants),
    ),
  );

  if (syncSheetError) {
    console.error(
      "❌ Error syncing product variants to sheet:",
      syncSheetError,
    );
    throw syncSheetError;
  }

  console.info("✅ product variants synced back to sheet");
  console.info(
    `Sheet ${env.GOOGLE_PROD_VARIANTS_SHEET_ID} updated successfully`,
  );

  return result;
};

export const syncProductVariants = (inngest: Inngest) => {
  return inngest.createFunction(
    {
      id: "sync-product-variants",
      retries: 3,
    },
    {
      cron: "*/30 * * * *",
    },
    syncProductVariantsFunc,
  );
};
