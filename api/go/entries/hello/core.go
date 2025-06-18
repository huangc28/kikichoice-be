package handler

import (
	"net/http"

	"github.com/huangc28/kikichoice-be/api/go/_internal/handlers/hello"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/logger"
	router "github.com/huangc28/kikichoice-be/api/go/_internal/router"
	routerfx "github.com/huangc28/kikichoice-be/api/go/_internal/router/fx"

	"github.com/go-chi/chi/v5"
	appfx "github.com/huangc28/kikichoice-be/api/go/_internal/fx"
	"go.uber.org/fx"
)

// Handle is the main entry point for the hello endpoint
// This demonstrates the basic pattern for creating Vercel serverless function handlers
func Handle(w http.ResponseWriter, r *http.Request) {
	fx.New(
		logger.TagLogger("hello"),
		appfx.CoreConfigOptions,
		routerfx.CoreRouterOptions,

		fx.Provide(
			router.AsRoute(hello.NewHelloHandler),
		),

		fx.Invoke(func(router *chi.Mux) {
			router.ServeHTTP(w, r)
		}),
	)
}
