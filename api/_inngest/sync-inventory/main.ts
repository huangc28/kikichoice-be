import { Inngest, NonRetriableError } from "inngest";
import type { GetFunctionInput } from "inngest";
import { tryCatch } from "#shared/try-catch.js";
import { upsertProducts } from "./upsert-products.js";
import { fetchSheetData } from "./services/google-sheet/fetch-sheet-data.js";
import { syncSheetCurrentStock } from "./services/google-sheet/sync-sheet-current-stock.js";

export const syncFunc = async ({ step }: GetFunctionInput<Inngest>) => {
  const [products, fetchError] = await tryCatch(
    step.run("fetch-sheet-data", fetchSheetData),
  );

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

  const [, syncError] = await tryCatch(
    step.run(
      "sync-sheet-current-stock",
      async () =>
        syncSheetCurrentStock(
          result!.updatedProducts.map((p) => ({
            sku: p.sku,
            stock: p.stock_count,
          })),
        ),
    ),
  );
  if (syncError) {
    console.error("Error clearing sheet adjustment:", syncError);
    throw syncError;
  }

  console.log(
    `✅ Successfully sync sheet stock, ${result?.updatedProducts.length}`,
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
