import client from "#shared/db.js";
import { generateSqlBatchParams } from "#shared/sql-batch-utils.js";
import type { ProductRow, ProductWithUUID } from "./types.js";
import { nanoid } from "nanoid";

export const upsertProducts = async (products: ProductRow[]): Promise<{
  inserted: number;
  updated: number;
  total: number;
  updatedProducts: Array<{ sku: string; stock_count: number }>;
}> => {
  if (products.length === 0) {
    console.info("No products to update");
    return { inserted: 0, updated: 0, total: 0, updatedProducts: [] };
  }

  // Deduplicate products by SKU to avoid conflict errors
  // First, track which SKUs appear multiple times
  const skuCounts = new Map<string, number>();
  const duplicateSkus: string[] = [];

  products.forEach((product) => {
    const count = (skuCounts.get(product.sku) || 0) + 1;
    skuCounts.set(product.sku, count);

    if (count === 2) { // Mark as duplicate when we see it for the second time
      duplicateSkus.push(product.sku);
    }
  });

  // Now deduplicate by keeping the last occurrence of each SKU
  const uniqueProducts = products.reduce((acc, product) => {
    acc.set(product.sku, product); // This will keep the last occurrence of each SKU
    return acc;
  }, new Map<string, ProductRow>());

  const deduplicatedProducts = Array.from(uniqueProducts.values());

  if (deduplicatedProducts.length !== products.length) {
    const duplicateCount = products.length -
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
    const allUpdatedProducts: Array<{ sku: string; stock_count: number }> = [];

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
      allUpdatedProducts.push(...result.updatedProducts);
    }

    const finalResult = {
      inserted: totalInserted,
      updated: totalUpdated,
      total: totalProcessed,
      updatedProducts: allUpdatedProducts,
    };

    console.log("✅ Batch upsert completed:", finalResult);
    return finalResult;
  } catch (error) {
    console.error("❌ Error during batch upsert:", error);
    throw new Error(`Database upsert failed: ${error}`);
  }
};

/**
 * Convenience function for common product upsert pattern
 */
export const generateProductBatchParams = (
  products: Array<ProductWithUUID>,
) => {
  return generateSqlBatchParams({
    records: products,
    valueExtractor: (product) => [
      product.uuid,
      product.sku,
      product.name,
      product.ready_for_sale,
      product.stock_adjust_count,
      product.price,
      product.short_desc,
    ],
    additionalExpressions: ["NOW()"],
  });
};

// Process a single batch of products
const processBatch = async (products: ProductRow[]): Promise<{
  inserted: number;
  updated: number;
  total: number;
  updatedProducts: Array<{ sku: string; stock_count: number }>;
}> => {
  const productsWithUuids = products.map((product) => ({
    ...product,
    uuid: nanoid(),
  }));

  const { valuesClause, values } = generateProductBatchParams(
    productsWithUuids,
  );

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
        stock_count = products.stock_count + EXCLUDED.stock_count,
        price = EXCLUDED.price,
        short_desc = EXCLUDED.short_desc,
        updated_at = NOW()
      RETURNING
        sku,
        stock_count,
        (xmax = 0) AS inserted
    )
    SELECT
      sku,
      stock_count,
      inserted
    FROM upsert_result;
  `;

  const result = await client.query(query, values);
  const rows = result.rows;

  // Calculate statistics from the returned rows
  const inserted = rows.filter((row) => row.inserted === true).length;
  const updated = rows.filter((row) => row.inserted === false).length;
  const total = rows.length;

  // Extract updated products info
  const updatedProducts = rows.map((row) => ({
    sku: row.sku,
    stock_count: parseInt(row.stock_count),
  }));

  return {
    inserted,
    updated,
    total,
    updatedProducts,
  };
};
