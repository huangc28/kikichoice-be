package hello

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/render"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HelloResponse represents the response structure for hello endpoint
type HelloResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
}

// HelloHandler handles hello world requests
type HelloHandler struct {
	logger *zap.SugaredLogger
}

// HelloHandlerParams defines dependencies for the hello handler
type HelloHandlerParams struct {
	fx.In

	Logger *zap.SugaredLogger
}

// NewHelloHandler creates a new hello handler instance
func NewHelloHandler(p HelloHandlerParams) *HelloHandler {
	return &HelloHandler{
		logger: p.Logger,
	}
}

// RegisterRoutes registers the hello routes with the chi router
func (h *HelloHandler) RegisterRoutes(r *chi.Mux) {
	r.Get("/hello", h.Handle)
}

// Handle processes the basic hello request
func (h *HelloHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing hello request")

	response := HelloResponse{
		Message:   "Hello, World! Welcome to the Go Chi Vercel Starter!",
		Timestamp: time.Now(),
		Path:      r.URL.Path,
		Method:    r.Method,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorw("Failed to encode response", "error", err)
		render.ChiErr(
			w, r, err, FailedToEncodeResponse,
			render.WithStatusCode(http.StatusInternalServerError),
		)
		return
	}

	render.ChiJSON(w, r, response)
}
