-- Rename table from product_variant to product_variants
ALTER TABLE product_variant RENAME TO product_variants;

-- Add price column to product_variants table
ALTER TABLE product_variants
ADD COLUMN price DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (price >= 0);

-- Update indexes to reflect new table name
DROP INDEX IF EXISTS idx_product_variant_product_id;
DROP INDEX IF EXISTS idx_product_variant_sku;
DROP INDEX IF EXISTS idx_product_variant_name;

CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_sku ON product_variants(sku);
CREATE INDEX idx_product_variants_name ON product_variants(name);
CREATE INDEX idx_product_variants_price ON product_variants(price);
