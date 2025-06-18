package handler

import (
	"net/http"

	appfx "github.com/huangc28/kikichoice-be/api/go/_internal/fx"
	"github.com/huangc28/kikichoice-be/api/go/_internal/handlers/products"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/logger"
	router "github.com/huangc28/kikichoice-be/api/go/_internal/router"
	routerfx "github.com/huangc28/kikichoice-be/api/go/_internal/router/fx"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	fx.New(
		logger.TagLogger("products"),
		appfx.CoreConfigOptions,
		routerfx.CoreRouterOptions,
		fx.Provide(
			products.NewProductListDAO,
		),
		fx.Provide(
			router.AsRoute(products.NewProductsListHandler),
		),
		fx.Invoke(func(router *chi.Mux) {
			router.ServeHTTP(w, r)
		}),
	)
}
