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

  try {
    console.log(`🚀 Starting batch upsert for ${products.length} products...`);

    // Split into smaller batches to avoid parameter limits
    const batchSize = 100; // PostgreSQL parameter limit consideration
    const batches: ProductRow[][] = [];

    for (let i = 0; i < products.length; i += batchSize) {
      batches.push(products.slice(i, i + batchSize));
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
      ON CONFLICT (uuid)
      DO UPDATE SET
        sku = EXCLUDED.sku,
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
