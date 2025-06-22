package products

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/render"
	router "github.com/huangc28/kikichoice-be/api/go/_internal/router"
)

type ProductVariantsListHandler struct {
	dao *ProductDAO
}

func NewProductVariantsListHandler(dao *ProductDAO) *ProductVariantsListHandler {
	return &ProductVariantsListHandler{dao: dao}
}

func (h *ProductVariantsListHandler) RegisterRoutes(r *chi.Mux) {
	r.Get("/v1/products/{uuid}/variants", h.Handle)
}

func (h *ProductVariantsListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")

	variants, err := h.dao.GetProductVariantsByUUID(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			render.ChiJSON(
				w, r,
				[]ProductVariantWithImage{},
			)
			return
		}
		// This will only trigger if:
		// 1. Product doesn't exist
		// 2. Product is not ready_for_sale
		// 3. Database error occurred
		render.ChiErr(
			w, r,
			err,
			GetProductVariantsFailed,
			render.WithStatusCode(http.StatusNotFound),
		)
		return
	}

	// Always return a response, even if variants slice is empty
	response := renderProductVariantsList(variants)
	render.ChiJSON(w, r, response)
}

var _ router.Handler = (*ProductVariantsListHandler)(nil)
