import { Inngest, NonRetriableError } from "inngest";
import type { GetFunctionInput } from "inngest";
import { tryCatch } from "#shared/try-catch.js";

import { env } from "#shared/env.js";
import { fetchProductVariantsFromSheet } from "./services/fetch-product-variants.js";
import { syncProductVariants as upsertProductVariants } from "./services/upsert-product-variants.js";
import { syncProductVariantsSheet } from "./services/sync-product-variants-sheet.js";
import { syncParentProductsStock } from "./services/sync-parent-products-stock.js";
import { updateParentProductsStockInDB } from "./services/update-parent-products-db.js";

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

  // Sync parent products stock totals to the parent sheet
  const [, syncParentError] = await tryCatch(
    step.run(
      "sync-parent-products-stock",
      () => syncParentProductsStock(result!.processedVariants),
    ),
  );

  if (syncParentError) {
    console.error(
      "❌ Error syncing parent products stock:",
      syncParentError,
    );
    throw syncParentError;
  }

  console.info("✅ parent products stock synced to sheet");
  console.info(
    `Parent sheet ${env.GOOGLE_SHEET_ID} updated successfully`,
  );

  // Update parent products stock in database
  const [, updateDbError] = await tryCatch(
    step.run(
      "update-parent-products-db",
      () => updateParentProductsStockInDB(result!.processedVariants),
    ),
  );

  if (updateDbError) {
    console.error(
      "❌ Error updating parent products in database:",
      updateDbError,
    );
    throw updateDbError;
  }

  console.info("✅ parent products stock updated in database");

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
