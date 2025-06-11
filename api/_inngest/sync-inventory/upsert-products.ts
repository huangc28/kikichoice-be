import type { ProductRow } from "./fetch-sheet-data.ts";
import client from "#shared/db.js";
import { generateProductBatchParams } from "#shared/sql-batch-utils.js";

export const upsertProducts = async (products: ProductRow[]): Promise<{
  inserted: number;
  updated: number;
  total: number;
}> => {
  if (products.length === 0) {
    console.info("No products to update");
    return { inserted: 0, updated: 0, total: 0 };
  }

  // Import nanoid dynamically since it's an ES module
  const { nanoid } = await import("nanoid");

  // Generate UUIDs for products that don't have them
  const productsWithUuids = products.map((product) => ({
    ...product,
    uuid: product.uuid || nanoid(),
  }));

  console.info("productsWithUuids", productsWithUuids);

  // Deduplicate products by SKU to avoid conflict errors
  // First, track which SKUs appear multiple times
  const skuCounts = new Map<string, number>();
  const duplicateSkus: string[] = [];

  productsWithUuids.forEach((product) => {
    const count = (skuCounts.get(product.sku) || 0) + 1;
    skuCounts.set(product.sku, count);

    if (count === 2) { // Mark as duplicate when we see it for the second time
      duplicateSkus.push(product.sku);
    }
  });

  // Now deduplicate by keeping the last occurrence of each SKU
  const uniqueProducts = productsWithUuids.reduce((acc, product) => {
    acc.set(product.sku, product); // This will keep the last occurrence of each SKU
    return acc;
  }, new Map<string, ProductRow>());

  const deduplicatedProducts = Array.from(uniqueProducts.values());

  if (deduplicatedProducts.length !== productsWithUuids.length) {
    const duplicateCount = productsWithUuids.length -
      deduplicatedProducts.length;
    console.warn(
      `⚠️ Found ${duplicateCount} duplicate SKUs, processing ${deduplicatedProducts.length} unique products`,
    );
    console.warn(`🔍 Duplicate SKUs found: ${duplicateSkus.join(", ")}`);
  }

  try {
    console.log(
      `🚀 Starting batch upsert for ${deduplicatedProducts.length} products...`,
    );

    // Split into smaller batches to avoid parameter limits
    const batchSize = 100; // PostgreSQL parameter limit consideration
    const batches: ProductRow[][] = [];

    for (let i = 0; i < deduplicatedProducts.length; i += batchSize) {
      batches.push(deduplicatedProducts.slice(i, i + batchSize));
    }

    let totalInserted = 0;
    let totalUpdated = 0;
    let totalProcessed = 0;

    // Process each batch
    for (const [batchIndex, batch] of batches.entries()) {
      console.log(
        `📦 Processing batch ${
          batchIndex + 1
        }/${batches.length} (${batch.length} products)`,
      );

      const result = await processBatch(batch);

      totalInserted += result.inserted;
      totalUpdated += result.updated;
      totalProcessed += result.total;
    }

    const finalResult = {
      inserted: totalInserted,
      updated: totalUpdated,
      total: totalProcessed,
    };

    console.log("✅ Batch upsert completed:", finalResult);
    return finalResult;
  } catch (error) {
    console.error("❌ Error during batch upsert:", error);
    throw new Error(`Database upsert failed: ${error}`);
  }
};

// Process a single batch of products
const processBatch = async (products: ProductRow[]): Promise<{
  inserted: number;
  updated: number;
  total: number;
}> => {
  const { valuesClause, values } = generateProductBatchParams(products);

  const query = `
    WITH upsert_result AS (
      INSERT INTO products (
        uuid,
        sku,
        name,
        ready_for_sale,
        stock_count,
        price,
        short_desc,
        updated_at
      )
      VALUES ${valuesClause}
      ON CONFLICT (sku)
      DO UPDATE SET
        name = EXCLUDED.name,
        ready_for_sale = EXCLUDED.ready_for_sale,
        stock_count = EXCLUDED.stock_count,
        price = EXCLUDED.price,
        short_desc = EXCLUDED.short_desc,
        updated_at = NOW()
      RETURNING
        uuid,
        (xmax = 0) AS inserted
    )
    SELECT
      COUNT(*) FILTER (WHERE inserted = true) as inserted_count,
      COUNT(*) FILTER (WHERE inserted = false) as updated_count,
      COUNT(*) as total_count
    FROM upsert_result;
  `;

  const result = await client.query(query, values);
  const stats = result.rows[0];

  return {
    inserted: parseInt(stats.inserted_count || 0),
    updated: parseInt(stats.updated_count || 0),
    total: parseInt(stats.total_count || 0),
  };
};
