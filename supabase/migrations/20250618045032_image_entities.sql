-- Create entity_type enum
CREATE TYPE entity_type AS ENUM ('product', 'product_variant');

-- Rename product_images table to image_entities
ALTER TABLE product_images RENAME TO image_entities;

-- Rename product_id column to entity_id
ALTER TABLE image_entities RENAME COLUMN product_id TO entity_id;

-- Add image_id column (referencing the images table)
ALTER TABLE image_entities ADD COLUMN image_id BIGINT NOT NULL REFERENCES images(id) ON DELETE CASCADE;

-- Add entity_type column
ALTER TABLE image_entities ADD COLUMN entity_type entity_type NOT NULL DEFAULT 'product';

-- Rename sequence to match new table name
ALTER SEQUENCE product_images_id_seq RENAME TO image_entities_id_seq;

-- Update sequence ownership
ALTER SEQUENCE image_entities_id_seq OWNED BY image_entities.id;

-- Drop old indexes
DROP INDEX IF EXISTS idx_product_images_is_primary;
DROP INDEX IF EXISTS idx_product_images_product_id;

-- Create new indexes with updated names
CREATE INDEX idx_image_entities_is_primary ON image_entities(is_primary);
CREATE INDEX idx_image_entities_entity_id ON image_entities(entity_id);
CREATE INDEX idx_image_entities_image_id ON image_entities(image_id);
CREATE INDEX idx_image_entities_entity_type ON image_entities(entity_type);

-- Drop old foreign key constraint
ALTER TABLE image_entities DROP CONSTRAINT IF EXISTS product_images_product_id_fkey;

-- Since we're making this more generic, we'll remove the specific foreign key to products
-- Applications will need to handle referential integrity based on entity_type
-- Alternatively, you could create conditional constraints or use triggers

-- Update table permissions (copy from old table)
GRANT DELETE ON TABLE image_entities TO anon;
GRANT INSERT ON TABLE image_entities TO anon;
GRANT REFERENCES ON TABLE image_entities TO anon;
GRANT SELECT ON TABLE image_entities TO anon;
GRANT TRIGGER ON TABLE image_entities TO anon;
GRANT TRUNCATE ON TABLE image_entities TO anon;
GRANT UPDATE ON TABLE image_entities TO anon;

GRANT DELETE ON TABLE image_entities TO authenticated;
GRANT INSERT ON TABLE image_entities TO authenticated;
GRANT REFERENCES ON TABLE image_entities TO authenticated;
GRANT SELECT ON TABLE image_entities TO authenticated;
GRANT TRIGGER ON TABLE image_entities TO authenticated;
GRANT TRUNCATE ON TABLE image_entities TO authenticated;
GRANT UPDATE ON TABLE image_entities TO authenticated;

GRANT DELETE ON TABLE image_entities TO service_role;
GRANT INSERT ON TABLE image_entities TO service_role;
GRANT REFERENCES ON TABLE image_entities TO service_role;
GRANT SELECT ON TABLE image_entities TO service_role;
GRANT TRIGGER ON TABLE image_entities TO service_role;
GRANT TRUNCATE ON TABLE image_entities TO service_role;
GRANT UPDATE ON TABLE image_entities TO service_role;
