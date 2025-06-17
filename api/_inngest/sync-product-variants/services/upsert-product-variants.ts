import dbpool from "#shared/db.js";
import { generateSqlBatchParams } from "#shared/sql-batch-utils.js";
import { ProductVariant } from "../types.js";

type ParentProd = {
  id: string;
  sku: string;
  price: number;
};

const fetchParentProdsByParentSkus = async (
  parentSkus: string[],
): Promise<ParentProd[]> => {
  if (parentSkus.length === 0) {
    return [];
  }

  // Fix SQL injection by using ANY() with parameterized query
  const query = `
    SELECT id, sku, price FROM products
    WHERE sku = ANY($1);
  `;

  const { rows } = await dbpool.query(query, [parentSkus]);
  return rows as ParentProd[];
};

export type ProcessedProductVariant = {
  sku: string;
  stock_count: number;
  price: number;
};

export const syncProductVariants = async (
  productVariants: ProductVariant[],
) => {
  if (productVariants.length === 0) {
    console.info("No product variants to sync");
    return {
      inserted: 0,
      updated: 0,
      total: 0,
      skipped: 0,
      processedVariants: [] as ProcessedProductVariant[],
    };
  }

  const parentProds = await fetchParentProdsByParentSkus(
    productVariants.map((variant) => variant.parent_sku),
  );

  const parentSkuToData = new Map(
    parentProds.map((prod) => [prod.sku, { id: prod.id, price: prod.price }]),
  );

  console.info("Found parent products:", parentProds.length);

  const validVariants = productVariants.filter((variant) => {
    const hasParent = parentSkuToData.has(variant.parent_sku);
    if (!hasParent) {
      console.warn(
        `⚠️ Parent product not found for SKU: ${variant.parent_sku}`,
      );
    }
    return hasParent;
  });

  if (validVariants.length === 0) {
    console.warn(
      "No valid product variants to sync (no matching parent products)",
    );
    return {
      inserted: 0,
      updated: 0,
      total: 0,
      skipped: productVariants.length,
      processedVariants: [] as ProcessedProductVariant[],
    };
  }

  console.info(
    `Processing ${validVariants.length} valid variants out of ${productVariants.length} total`,
  );

  // Prepare data for batch upsert
  const variantsWithProductId = validVariants.map((variant) => {
    const parentData = parentSkuToData.get(variant.parent_sku)!;
    return {
      product_id: parentData.id,
      name: variant.name,
      sku: variant.sku,
      stock_count: variant.stock_adjust_count,
      price: variant.price === 0 ? parentData.price : variant.price,
    };
  });

  // Use batch processing for large datasets
  const batchSize = 100;
  const batches: typeof variantsWithProductId[] = [];

  for (let i = 0; i < variantsWithProductId.length; i += batchSize) {
    batches.push(variantsWithProductId.slice(i, i + batchSize));
  }

  let totalInserted = 0;
  let totalUpdated = 0;
  let totalProcessed = 0;
  const allProcessedVariants: ProcessedProductVariant[] = [];

  for (const [batchIndex, batch] of batches.entries()) {
    console.log(
      `📦 Processing batch ${
        batchIndex + 1
      }/${batches.length} (${batch.length} variants)`,
    );

    const result = await processBatch(batch);
    totalInserted += result.inserted;
    totalUpdated += result.updated;
    totalProcessed += result.total;
    allProcessedVariants.push(...result.processedVariants);
  }

  const finalResult = {
    inserted: totalInserted,
    updated: totalUpdated,
    total: totalProcessed,
    skipped: productVariants.length - validVariants.length,
    processedVariants: allProcessedVariants,
  };

  console.log("✅ Product variants sync completed:", finalResult);
  return finalResult;
};

const processBatch = async (
  variants: Array<{
    product_id: string;
    name: string;
    sku: string;
    stock_count: number;
    price: number;
  }>,
): Promise<{
  inserted: number;
  updated: number;
  total: number;
  processedVariants: ProcessedProductVariant[];
}> => {
  const { valuesClause, values } = generateSqlBatchParams({
    records: variants,
    valueExtractor: (variant) => [
      variant.product_id,
      variant.name,
      variant.sku,
      variant.stock_count,
      variant.price,
    ],
    additionalExpressions: ["NOW()", "NOW()"],
  });

  const query = `
    WITH upsert_result AS (
      INSERT INTO product_variants (
        product_id,
        name,
        sku,
        stock_count,
        price,
        created_at,
        updated_at
      )
      VALUES ${valuesClause}
      ON CONFLICT (sku)
      DO UPDATE SET
        product_id = EXCLUDED.product_id,
        name = EXCLUDED.name,
        stock_count = product_variants.stock_count + EXCLUDED.stock_count,
        price = EXCLUDED.price,
        updated_at = NOW()
      RETURNING
        sku,
        stock_count,
        price,
        (xmax = 0) AS inserted
    )
    SELECT
      sku,
      stock_count,
      price,
      inserted
    FROM upsert_result;
  `;

  const result = await dbpool.query(query, values);
  const rows = result.rows;

  // Calculate statistics from the returned rows
  const inserted = rows.filter((row) => row.inserted === true).length;
  const updated = rows.filter((row) => row.inserted === false).length;
  const total = rows.length;

  // Extract processed variants data for sheet sync
  const processedVariants: ProcessedProductVariant[] = rows.map((row) => ({
    sku: row.sku,
    stock_count: parseInt(row.stock_count),
    price: parseFloat(row.price),
  }));

  return {
    inserted,
    updated,
    total,
    processedVariants,
  };
};
