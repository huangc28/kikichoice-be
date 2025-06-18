package router

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	wvmiddlewares "github.com/huangc28/kikichoice-be/api/go/_internal/middlewares"
	"go.uber.org/fx"
)

type Handler interface {
	RegisterRoutes(r *chi.Mux)
	Handle(w http.ResponseWriter, r *http.Request)
}

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Handler)),
		fx.ResultTags(`group:"handlers"`),
	)
}

func NewRouter(handlers []Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(wvmiddlewares.Logger)
	for _, handler := range handlers {
		log.Printf("registering handler: %T", handler)
		handler.RegisterRoutes(r)
	}
	return r
}
