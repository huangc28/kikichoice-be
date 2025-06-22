package products

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/render"
	router "github.com/huangc28/kikichoice-be/api/go/_internal/router"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type HotSellingHandler struct {
	dao    *ProductDAO
	logger *zap.SugaredLogger
}

type HotSellingHandlerParams struct {
	fx.In

	DAO    *ProductDAO
	Logger *zap.SugaredLogger
}

func NewHotSellingHandler(p HotSellingHandlerParams) *HotSellingHandler {
	return &HotSellingHandler{
		dao:    p.DAO,
		logger: p.Logger,
	}
}

func (h *HotSellingHandler) RegisterRoutes(r *chi.Mux) {
	r.Get("/v1/products/hot-selling", h.Handle)
}

func (h *HotSellingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	products, err := h.dao.GetHotSellingProducts(ctx)
	if err != nil {
		h.logger.Errorw("Failed to get hot selling products", "error", err)
		render.ChiErr(
			w, r,
			err,
			GetProductsFailed,
			render.WithStatusCode(http.StatusInternalServerError),
		)
		return
	}

	response := renderProductList(products)
	render.ChiJSON(w, r, response)
}

var _ router.Handler = (*HotSellingHandler)(nil)
