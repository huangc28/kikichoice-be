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
	Slug            string         `json:"slug"`
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
			Slug:            product.Slug.String,
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

func renderProductDetail(productDetail *ProductDetail) *ProductDetailResponse {
	// Convert specs from JSON format
	specs := make([]ProductSpecResponse, len(productDetail.ParsedSpecs))
	for i, spec := range productDetail.ParsedSpecs {
		specs[i] = ProductSpecResponse{
			Name:  spec.SpecName,
			Value: spec.SpecValue,
		}
	}

	// Convert images
	images := make([]ProductImageResponse, len(productDetail.Images))
	for i, image := range productDetail.Images {
		images[i] = ProductImageResponse{
			URL:       image.URL,
			IsPrimary: image.IsPrimary,
		}
	}

	// Convert variants
	variants := make([]ProductVariantResponse, len(productDetail.Variants))
	for i, variant := range productDetail.Variants {
		variants[i] = ProductVariantResponse{
			Name:       variant.Name,
			SKU:        variant.Sku,
			StockCount: variant.StockCount,
			ImageURL:   variant.ImageURL.String,
			Price:      variant.Price,
			UUID:       variant.Uuid.String,
		}
	}

	return &ProductDetailResponse{
		UUID:          productDetail.Product.Uuid,
		SKU:           productDetail.Product.Sku,
		Name:          productDetail.Product.Name,
		Slug:          productDetail.Product.Slug.String,
		Price:         productDetail.Product.Price,
		OriginalPrice: productDetail.Product.OriginalPrice,
		ShortDesc:     productDetail.Product.ShortDesc.String,
		FullDesc:      productDetail.Product.FullDesc.String,
		StockCount:    productDetail.Product.StockCount,
		Images:        images,
		Specs:         specs,
		Variants:      variants,
	}
}
