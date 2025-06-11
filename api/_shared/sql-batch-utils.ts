/**
 * Configuration for SQL batch operations
 */
export interface SqlBatchConfig<T> {
  /** Array of records to process */
  records: T[];
  /** Function to extract values from each record */
  valueExtractor: (record: T) => any[];
  /** Additional SQL expressions to append to each row (e.g., 'NOW()', 'DEFAULT') */
  additionalExpressions?: string[];
}

/**
 * Result of SQL batch parameter generation
 */
export interface SqlBatchResult {
  /** SQL VALUES clause with parameter placeholders */
  valuesClause: string;
  /** Flattened array of parameter values */
  values: any[];
  /** Number of parameters per record */
  parametersPerRecord: number;
}

/**
 * Generates SQL parameter placeholders and flattened values for batch operations
 *
 * @example
 * ```typescript
 * const products = [
 *   { uuid: '1', name: 'Product 1', price: 10.99 },
 *   { uuid: '2', name: 'Product 2', price: 20.99 }
 * ];
 *
 * const result = generateSqlBatchParams({
 *   records: products,
 *   valueExtractor: (product) => [product.uuid, product.name, product.price],
 *   additionalExpressions: ['NOW()']
 * });
 *
 * // Result:
 * // valuesClause: "($1, $2, $3, NOW()), ($4, $5, $6, NOW())"
 * // values: ['1', 'Product 1', 10.99, '2', 'Product 2', 20.99]
 * ```
 */
export const generateSqlBatchParams = <T>(
  config: SqlBatchConfig<T>,
): SqlBatchResult => {
  const { records, valueExtractor, additionalExpressions = [] } = config;

  if (records.length === 0) {
    return {
      valuesClause: "",
      values: [],
      parametersPerRecord: 0,
    };
  }

  // Extract values from first record to determine parameter count
  const firstRecordValues = valueExtractor(records[0]);
  const parametersPerRecord = firstRecordValues.length;

  // Generate parameter placeholders for each record
  const valuesClause = records.map((_, index) => {
    const baseIndex = index * parametersPerRecord;

    // Generate parameter placeholders ($1, $2, $3, etc.)
    const paramPlaceholders = Array.from(
      { length: parametersPerRecord },
      (_, i) => `$${baseIndex + i + 1}`,
    );

    // Combine with additional expressions
    const allExpressions = [...paramPlaceholders, ...additionalExpressions];

    return `(${allExpressions.join(", ")})`;
  }).join(", ");

  // Flatten all values into a single array
  const values = records.flatMap(valueExtractor);

  return {
    valuesClause,
    values,
    parametersPerRecord,
  };
};

/**
 * Convenience function for common product upsert pattern
 */
export const generateProductBatchParams = (
  products: Array<{
    sku: string;
    name: string;
    ready_for_sale: boolean;
    stock_count: number;
    price: number;
    short_desc: string;
  }>,
) => {
  return generateSqlBatchParams({
    records: products,
    valueExtractor: (product) => [
      product.sku,
      product.name,
      product.ready_for_sale,
      product.stock_count,
      product.price,
      product.short_desc,
    ],
    additionalExpressions: ["NOW()"],
  });
};
