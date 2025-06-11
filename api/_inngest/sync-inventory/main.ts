import { Inngest, NonRetriableError } from "inngest";
import type { GetFunctionInput } from "inngest";
import { tryCatch } from "#shared/try-catch.js";
import { upsertProducts } from "./upsert-products.js";
import { fetchSheetData } from "./services/google-sheet/fetch-sheet-data.js";
import { clearSheetAdjustment } from "./services/google-sheet/clear-sheet_adjustment.js";

export const syncFunc = async ({ step }: GetFunctionInput<Inngest>) => {
  const [products, fetchError] = await tryCatch(
    step.run("fetch-sheet-data", async () => fetchSheetData()),
  );

  console.log("** 1 products", products);

  if (fetchError) {
    console.error("Error fetching sheet data:", fetchError);
    throw fetchError;
  }

  if (!products || products.length === 0) {
    console.info("No products to update");
    throw new NonRetriableError("No products to update");
  }

  const [result, upsertError] = await tryCatch(
    step.run("upsert-products", async () => upsertProducts(products)),
  );

  console.info("result synced", result);

  if (upsertError) {
    console.error("Error upserting products:", upsertError);
    throw upsertError;
  }

  const productWithAdjustment = products.filter(
    (p) => p.stock_adjust_count > 0,
  );

  const [, clearError] = await tryCatch(
    step.run(
      "clear-sheet-adjustment",
      async () => clearSheetAdjustment(productWithAdjustment.map((p) => p.sku)),
    ),
  );
  if (clearError) {
    console.error("Error clearing sheet adjustment:", clearError);
    throw clearError;
  }

  console.log(
    `✅ Successfully cleared sheet adjustment, ${productWithAdjustment.length}`,
  );

  return {
    inserted: result?.inserted || 0,
    updated: result?.updated || 0,
    total: result?.total || 0,
  };
};

export const syncInventory = (inngest: Inngest) => {
  return inngest.createFunction(
    {
      id: "sync-inventory",
      retries: 3,
    },
    { cron: "*/30 * * * *" }, // every 30 minutes
    syncFunc,
  );
};
