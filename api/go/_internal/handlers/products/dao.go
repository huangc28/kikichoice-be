package products

import (
	"context"

	"github.com/huangc28/kikichoice-be/api/go/_internal/db"
)

type ProductListDAO struct {
	db db.Conn
}

func NewProductListDAO(db db.Conn) *ProductListDAO {
	return &ProductListDAO{db: db}
}

func (dao *ProductListDAO) GetProducts(ctx context.Context, page, perPage int) ([]*Product, error) {
	offset := (page - 1) * perPage

	// Execute the complex query to get products with pagination
	query := `
		SELECT
			p.id,
			p.uuid,
			p.sku,
			p.name,
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
		err := rows.StructScan(&product)
		if err != nil {
			return nil, err
		}

		product.HasVariant = variantCount > 0
		products = append(products, &product)
	}

	return products, nil
}
