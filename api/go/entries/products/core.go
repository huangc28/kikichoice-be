package handler

import (
	"net/http"

	appfx "github/huangc28/kikichoice-be/api/go/_internal/fx"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/products"
	"github/huangc28/kikichoice-be/api/go/_internal/pkg/logger"
	router "github/huangc28/kikichoice-be/api/go/_internal/router"
	routerfx "github/huangc28/kikichoice-be/api/go/_internal/router/fx"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	fx.New(
		logger.TagLogger("products"),
		appfx.CoreConfigOptions,
		routerfx.CoreRouterOptions,
		fx.Provide(
			router.AsRoute(products.NewProductsListHandler),
		),
		fx.Invoke(func(router *chi.Mux) {
			router.ServeHTTP(w, r)
		}),
	)
}
