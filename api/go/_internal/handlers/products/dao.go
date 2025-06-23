package products

import (
	"context"
	"encoding/json"

	"github.com/huangc28/kikichoice-be/api/go/_internal/db"
)

type ProductDAO struct {
	db db.Conn
}

func NewProductDAO(db db.Conn) *ProductDAO {
	return &ProductDAO{db: db}
}

func (dao *ProductDAO) GetProducts(ctx context.Context, page, perPage int) ([]*Product, error) {
	offset := (page - 1) * perPage

	// Execute the complex query to get products with pagination
	query := `
		SELECT
			p.id,
			p.uuid,
			p.sku,
			p.name,
			p.slug,
			p.price,
			p.original_price,
			p.stock_count,
			p.short_desc,
			COALESCE(variant_count.count, 0) as variant_count,
			img.url as primary_image_url
		FROM products p
		LEFT JOIN (
			SELECT
				product_id,
				COUNT(*) as count
			FROM product_variants
			GROUP BY product_id
		) variant_count ON p.id = variant_count.product_id
		LEFT JOIN (
			SELECT DISTINCT ON (ie.entity_id)
				ie.entity_id,
				i.url
			FROM image_entities ie
			JOIN images i ON ie.image_id = i.id
			WHERE ie.entity_type = 'product' AND ie.is_primary = true
			ORDER BY ie.entity_id, ie.sort_order
		) img ON p.id = img.entity_id
		WHERE p.ready_for_sale = true
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := dao.db.Queryx(query, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*Product, 0)

	for rows.Next() {
		var product Product
		var variantCount int64

		// Manual scan to properly read all columns including variant_count
		err := rows.Scan(
			&product.ID,
			&product.Uuid,
			&product.Sku,
			&product.Name,
			&product.Slug,
			&product.Price,
			&product.OriginalPrice,
			&product.StockCount,
			&product.ShortDesc,
			&variantCount, // This was missing - now properly populated
			&product.PrimaryImageURL,
		)
		if err != nil {
			return nil, err
		}

		product.HasVariant = variantCount > 0
		products = append(products, &product)
	}

	return products, nil
}

func (dao *ProductDAO) GetProductByUUID(ctx context.Context, uuid string) (*ProductDetail, error) {
	// First, get the main product
	productQuery := `
		SELECT
			p.id,
			p.uuid,
			p.sku,
			p.name,
			p.slug,
			p.price,
			p.original_price,
			p.short_desc,
			p.full_desc,
			p.stock_count,
			p.specs,
			p.ready_for_sale,
			p.created_at,
			p.updated_at
		FROM products p
		WHERE p.uuid = $1 AND p.ready_for_sale = true
	`

	var product db.Product
	err := dao.db.Get(&product, productQuery, uuid)
	if err != nil {
		return nil, err
	}

	// Parse specs from JSONB column
	var specsJSON []ProductSpecJSON
	if len(product.Specs) > 0 {
		err = json.Unmarshal(product.Specs, &specsJSON)
		if err != nil {
			return nil, err
		}
	}

	// Get product images
	imagesQuery := `
		SELECT
			i.url,
			ie.is_primary,
			COALESCE(ie.sort_order, 0) as sort_order
		FROM image_entities ie
		JOIN images i ON ie.image_id = i.id
		WHERE ie.entity_id = $1 AND ie.entity_type = $2
		ORDER BY ie.sort_order, ie.id
	`

	var images []ProductImageWithEntity
	err = dao.db.Select(&images, imagesQuery, product.ID, db.EntityTypeProduct)
	if err != nil {
		return nil, err
	}

	// Get product variants with their primary images
	variantsQuery := `
		SELECT
			pv.id,
			pv.product_id,
			pv.name,
			pv.stock_count,
			pv.reserved_count,
			pv.sku,
			pv.price,
			pv.uuid,
			pv.created_at,
			pv.updated_at,
			COALESCE(img.url, '') as image_url
		FROM product_variants pv
		LEFT JOIN (
			SELECT DISTINCT ON (ie.entity_id)
				ie.entity_id,
				i.url
			FROM image_entities ie
			JOIN images i ON ie.image_id = i.id
			WHERE ie.entity_type = 'product_variant' AND ie.is_primary = true
			ORDER BY ie.entity_id, ie.sort_order
		) img ON pv.id = img.entity_id
		WHERE pv.product_id = $1
		ORDER BY pv.name
	`

	rows, err := dao.db.Queryx(variantsQuery, product.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []ProductVariantWithImage
	for rows.Next() {
		var variant ProductVariantWithImage
		err := rows.StructScan(&variant)
		if err != nil {
			return nil, err
		}
		variants = append(variants, variant)
	}

	return &ProductDetail{
		Product:     product,
		ParsedSpecs: specsJSON,
		Images:      images,
		Variants:    variants,
	}, nil
}

func (dao *ProductDAO) GetProductVariantsByUUID(ctx context.Context, productUUID string) ([]ProductVariantWithImage, error) {
	// First verify the product exists and is available for sale
	productQuery := `SELECT id FROM products WHERE uuid = $1 AND ready_for_sale = true`
	var productID int64
	err := dao.db.Get(&productID, productQuery, productUUID)
	if err != nil {
		// Return error only if product doesn't exist or database error
		return nil, err
	}

	// Get product variants with their primary images
	variantsQuery := `
		SELECT
			pv.id,
			pv.product_id,
			pv.name,
			pv.stock_count,
			pv.reserved_count,
			pv.sku,
			pv.price,
			pv.uuid,
			pv.created_at,
			pv.updated_at,
			COALESCE(img.url, '') as image_url
		FROM product_variants pv
		LEFT JOIN (
			SELECT DISTINCT ON (ie.entity_id)
				ie.entity_id,
				i.url
			FROM image_entities ie
			JOIN images i ON ie.image_id = i.id
			WHERE ie.entity_type = 'product_variant' AND ie.is_primary = true
			ORDER BY ie.entity_id, ie.sort_order
		) img ON pv.id = img.entity_id
		WHERE pv.product_id = $1
		ORDER BY pv.name
	`

	var variants []ProductVariantWithImage
	err = dao.db.Select(&variants, variantsQuery, productID)
	if err != nil {
		return nil, err
	}

	// Return empty slice if no variants found - this is not an error
	return variants, nil
}

func (dao *ProductDAO) GetHotSellingProducts(ctx context.Context) ([]*Product, error) {
	// Complex query that ranks products by sales in the past 30 days,
	// with fallback to created_at for products without sales data
	query := `
		WITH product_sales AS (
			SELECT
				oi.product_id,
				SUM(oi.quantity) as total_sold
			FROM order_items oi
			JOIN orders o ON oi.order_id = o.id
			WHERE o.created_at >= NOW() - INTERVAL '30 days'
				AND o.status IN ('paid', 'processing', 'shipped', 'delivered')
			GROUP BY oi.product_id
		),
		ranked_products AS (
			SELECT
				p.id,
				p.uuid,
				p.sku,
				p.name,
				p.slug,
				p.price,
				p.original_price,
				p.stock_count,
				p.short_desc,
				p.created_at,
				COALESCE(ps.total_sold, 0) as sales_count,
				COALESCE(variant_count.count, 0) as variant_count,
				img.url as primary_image_url,
				CASE
					WHEN ps.total_sold > 0 THEN 1
					ELSE 2
				END as sort_priority
			FROM products p
			LEFT JOIN product_sales ps ON p.id = ps.product_id
			LEFT JOIN (
				SELECT
					product_id,
					COUNT(*) as count
				FROM product_variants
				GROUP BY product_id
			) variant_count ON p.id = variant_count.product_id
			LEFT JOIN (
				SELECT DISTINCT ON (ie.entity_id)
					ie.entity_id,
					i.url
				FROM image_entities ie
				JOIN images i ON ie.image_id = i.id
				WHERE ie.entity_type = 'product' AND ie.is_primary = true
				ORDER BY ie.entity_id, ie.sort_order
			) img ON p.id = img.entity_id
			WHERE p.ready_for_sale = true
		)
		SELECT
			id,
			uuid,
			sku,
			name,
			slug,
			price,
			original_price,
			stock_count,
			short_desc,
			variant_count,
			primary_image_url
		FROM ranked_products
		ORDER BY
			sort_priority ASC,
			CASE WHEN sort_priority = 1 THEN sales_count END DESC,
			CASE WHEN sort_priority = 2 THEN created_at END DESC
		LIMIT 6
	`

	rows, err := dao.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*Product, 0)

	for rows.Next() {
		var product Product
		var variantCount int64

		// Manual scan to properly read all columns
		err := rows.Scan(
			&product.ID,
			&product.Uuid,
			&product.Sku,
			&product.Name,
			&product.Slug,
			&product.Price,
			&product.OriginalPrice,
			&product.StockCount,
			&product.ShortDesc,
			&variantCount,
			&product.PrimaryImageURL,
		)
		if err != nil {
			return nil, err
		}

		product.HasVariant = variantCount > 0
		product.VariantCount = variantCount
		products = append(products, &product)
	}

	return products, nil
}
