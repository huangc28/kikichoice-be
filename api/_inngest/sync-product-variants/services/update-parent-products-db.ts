import dbpool from "#shared/db.js";
import type { ProcessedProductVariant } from "./upsert-product-variants.js";

export const updateParentProductsStockInDB = async (
  processedVariants: ProcessedProductVariant[],
) => {
  if (processedVariants.length === 0) {
    console.info("No processed variants to update parent products stock in DB");
    return;
  }

  console.log(
    `🔄 Starting parent products stock update in database for ${processedVariants.length} variants...`,
  );

  // Calculate parent stock totals from variants
  const parentStockTotals = calculateParentStockTotals(processedVariants);
  console.info(
    `📈 Will update ${parentStockTotals.size} parent products in database`,
  );

  if (parentStockTotals.size === 0) {
    console.info("📭 No parent products to update in database");
    return;
  }

  // Update parent products in database
  const result = await updateParentProductsStock(parentStockTotals);
  console.info(
    `✅ Successfully updated ${result.updated} parent products in database`,
  );

  return result;
};

const calculateParentStockTotals = (
  processedVariants: ProcessedProductVariant[],
): Map<string, number> => {
  const parentStockMap = new Map<string, number>();

  processedVariants.forEach((variant) => {
    if (variant.parent_sku) {
      const currentTotal = parentStockMap.get(variant.parent_sku) || 0;
      parentStockMap.set(
        variant.parent_sku,
        currentTotal + variant.stock_count,
      );
    }
  });

  return parentStockMap;
};

const updateParentProductsStock = async (
  parentStockTotals: Map<string, number>,
): Promise<{ updated: number; total: number }> => {
  const parentSkus = Array.from(parentStockTotals.keys());
  const stockValues = Array.from(parentStockTotals.values());

  if (parentSkus.length === 0) {
    return { updated: 0, total: 0 };
  }

  // Create a CASE statement for bulk updates
  const caseStatements = parentSkus
    .map((sku, index) =>
      `WHEN sku = $${index + 2} THEN $${index + 2 + parentSkus.length}`
    )
    .join(" ");

  const query = `
    UPDATE products
    SET
      stock_count = CASE
        ${caseStatements}
        ELSE stock_count
      END,
      updated_at = NOW()
    WHERE sku = ANY($1)
    RETURNING sku, stock_count;
  `;

  const params = [
    parentSkus, // $1 - array of parent SKUs
    ...parentSkus, // $2 to $n - individual SKUs for CASE
    ...stockValues, // $n+1 to $2n - corresponding stock values
  ];

  const result = await dbpool.query(query, params);
  const updatedRows = result.rows;

  // Log the updates for debugging
  updatedRows.forEach((row) => {
    console.log(
      `📦 Database: Parent ${row.sku} stock updated to ${row.stock_count}`,
    );
  });

  return {
    updated: updatedRows.length,
    total: parentSkus.length,
  };
};
