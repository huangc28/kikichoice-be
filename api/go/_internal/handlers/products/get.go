package products

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/render"
	router "github.com/huangc28/kikichoice-be/api/go/_internal/router"
)

type ProductDetailHandler struct {
	dao *ProductDAO
}

func NewProductDetailHandler(dao *ProductDAO) *ProductDetailHandler {
	return &ProductDetailHandler{dao: dao}
}

func (h *ProductDetailHandler) RegisterRoutes(r *chi.Mux) {
	r.Get("/v1/products/{uuid}", h.Handle)
}

func (h *ProductDetailHandler) Handle(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")

	productDetail, err := h.dao.GetProductByUUID(r.Context(), uuid)
	if err != nil {
		render.ChiErr(
			w, r,
			err,
			GetProductFailed,
			render.WithStatusCode(http.StatusNotFound),
		)
		return
	}

	response := renderProductDetail(productDetail)
	render.ChiJSON(w, r, response)
}

var _ router.Handler = (*ProductDetailHandler)(nil)
