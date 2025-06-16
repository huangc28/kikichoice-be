package products

import "net/http"

type ProductsListHandler struct{}

func NewProductsListHandler() *ProductsListHandler {
	return &ProductsListHandler{}
}

func (h *ProductsListHandler) Handle(w http.ResponseWriter, r *http.Request) {}
