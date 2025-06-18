package products

import (
	"github.com/huangc28/kikichoice-be/api/go/_internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// ProductListResponse represents a single product in the list
type Product struct {
	db.Product
	PrimaryImageURL pgtype.Text `json:"primary_image_url"`
	HasVariant      bool        `json:"has_variant"`
	VariantCount    int64       `json:"variant_count"`
}
