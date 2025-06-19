package products

import (
	"net/http"
	"strconv"

	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/render"
	router "github.com/huangc28/kikichoice-be/api/go/_internal/router"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ProductsListHandler struct {
	dao       *ProductDAO
	validator *validator.Validate
	logger    *zap.SugaredLogger
}

type ProductsListHandlerParams struct {
	fx.In

	DAO    *ProductDAO
	Logger *zap.SugaredLogger
}

func NewProductsListHandler(p ProductsListHandlerParams) *ProductsListHandler {
	return &ProductsListHandler{
		dao:       p.DAO,
		validator: validator.New(),
		logger:    p.Logger,
	}
}

func (h *ProductsListHandler) RegisterRoutes(r *chi.Mux) {
	r.Get("/v1/products", h.Handle)
}

type ProductListQuery struct {
	Page    int `default:"1" validate:"required,min=1"`
	PerPage int `default:"15" validate:"required,min=1,max=100"`
}

func (h *ProductsListHandler) validateQuery(r *http.Request) (*ProductListQuery, error) {
	// Parse pagination parameters with defaults
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	// Default values
	page := 1
	perPage := 15

	// Parse page parameter if provided
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return nil, err
		}
	}

	// Parse perPage parameter if provided
	if perPageStr != "" {
		var err error
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			return nil, err
		}
	}

	query := &ProductListQuery{
		Page:    page,
		PerPage: perPage,
	}

	if err := h.validator.Struct(query); err != nil {
		return nil, err
	}

	return query, nil
}

func (h *ProductsListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query, err := h.validateQuery(r)
	if err != nil {
		render.ChiErr(
			w, r,
			err,
			InvalidQueryParams,
			render.WithStatusCode(http.StatusBadRequest),
		)
		return
	}

	products, err := h.dao.GetProducts(ctx, query.Page, query.PerPage)
	if err != nil {
		h.logger.Errorw("Failed to get products", "error", err)
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

var _ router.Handler = (*ProductsListHandler)(nil)
