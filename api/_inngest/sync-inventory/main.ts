import { Inngest, NonRetriableError } from "inngest";
import type { GetFunctionInput } from "inngest";
import { tryCatch } from "#shared/try-catch.js";
import { upsertProducts } from "./upsert-products.js";
import { fetchSheetData } from "./fetch-sheet-data.js";

export const syncFunc = async ({ step }: GetFunctionInput<Inngest>) => {
  const [products, fetchError] = await tryCatch(
    step.run("fetch-sheet-data", async () => fetchSheetData()),
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
