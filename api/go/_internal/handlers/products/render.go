package products

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// ProductListAPIResponse represents the complete API response
type ProductListAPIResponse struct {
	Products []*ProductResponse `json:"products"`
}

// ProductResponse represents a single product in the API response (without ID)
type ProductResponse struct {
	UUID            string         `json:"uuid"`
	SKU             string         `json:"sku"`
	Name            string         `json:"name"`
	Price           pgtype.Numeric `json:"price"`
	OriginalPrice   pgtype.Numeric `json:"original_price"`
	StockCount      int32          `json:"stock_count"`
	ShortDesc       pgtype.Text    `json:"short_desc"`
	VariantCount    int64          `json:"variant_count"`
	PrimaryImageURL pgtype.Text    `json:"primary_image_url"`
	HasVariant      bool           `json:"has_variant"`
}

func renderProductList(products []*Product) *ProductListAPIResponse {
	productResponses := make([]*ProductResponse, len(products))

	for i, product := range products {
		productResponses[i] = &ProductResponse{
			UUID:            product.Uuid,
			SKU:             product.Sku,
			Name:            product.Name,
			Price:           product.Price,
			OriginalPrice:   product.OriginalPrice,
			StockCount:      product.StockCount,
			ShortDesc:       product.ShortDesc,
			VariantCount:    product.VariantCount,
			PrimaryImageURL: product.PrimaryImageURL,
			HasVariant:      product.HasVariant,
		}
	}

	return &ProductListAPIResponse{
		Products: productResponses,
	}
}
