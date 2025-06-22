package products

import (
	"encoding/json"

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

// ProductDetailResponse represents the complete product detail API response
type ProductDetailResponse struct {
	UUID          string                   `json:"uuid"`
	SKU           string                   `json:"sku"`
	Name          string                   `json:"name"`
	Slug          string                   `json:"slug"`
	Price         pgtype.Numeric           `json:"price"`
	OriginalPrice pgtype.Numeric           `json:"original_price"`
	ShortDesc     string                   `json:"short_desc"`
	FullDesc      string                   `json:"full_desc"`
	StockCount    int32                    `json:"stock_count"`
	Images        []ProductImageResponse   `json:"images"`
	Specs         []ProductSpecResponse    `json:"specs"`
	Variants      []ProductVariantResponse `json:"variants"`
}

// ProductImageResponse represents a product image in the API response
type ProductImageResponse struct {
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
}

// ProductSpecResponse represents a product specification
type ProductSpecResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ProductSpecJSON represents the JSON structure of specs in the database
type ProductSpecJSON struct {
	SpecName  string `json:"spec_name"`
	SpecValue string `json:"spec_value"`
}

// ProductVariantResponse represents a product variant
type ProductVariantResponse struct {
	Name       string         `json:"name"`
	SKU        string         `json:"sku"`
	StockCount int32          `json:"stock_count"`
	ImageURL   string         `json:"image_url"`
	Price      pgtype.Numeric `json:"price"`
	UUID       string         `json:"uuid"`
}

// ProductVariantsListResponse represents the API response for product variants list
type ProductVariantsListResponse struct {
	Variants []ProductVariantResponse `json:"variants"`
}

// ProductDetail represents the internal product detail with all related data
type ProductDetail struct {
	db.Product
	ParsedSpecs []ProductSpecJSON         `json:"parsed_specs"`
	Images      []ProductImageWithEntity  `json:"images"`
	Variants    []ProductVariantWithImage `json:"variants"`
}

// ProductImageWithEntity represents a product image with entity metadata
type ProductImageWithEntity struct {
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

// ProductVariantWithImage represents a product variant with its primary image
type ProductVariantWithImage struct {
	db.ProductVariant
	ImageURL pgtype.Text `json:"image_url"`
}

// ParseSpecs parses the JSONB specs column into structured data
func (pd *ProductDetail) ParseSpecs() ([]ProductSpecJSON, error) {
	var specs []ProductSpecJSON
	if len(pd.Product.Specs) == 0 {
		return specs, nil
	}

	err := json.Unmarshal(pd.Product.Specs, &specs)
	if err != nil {
		return nil, err
	}

	return specs, nil
}
