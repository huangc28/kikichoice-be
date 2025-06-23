package webhooks

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/huangc28/kikichoice-be/api/go/_internal/pkg/render"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebhookHandler handles incoming webhooks
type WebhookHandler struct {
	userDAO *UserDAO
	logger  *zap.SugaredLogger
}

// WebhookHandlerParams defines dependencies for the webhook handler
type WebhookHandlerParams struct {
	fx.In

	UserDAO *UserDAO
	Logger  *zap.SugaredLogger
}

// NewWebhookHandler creates a new webhook handler instance
func NewWebhookHandler(p WebhookHandlerParams) *WebhookHandler {
	return &WebhookHandler{
		userDAO: p.UserDAO,
		logger:  p.Logger,
	}
}

// RegisterRoutes registers the webhook routes with the chi router
func (h *WebhookHandler) RegisterRoutes(r *chi.Mux) {
	r.Post("/v1/webhooks/clerk/create-user", h.Handle)
}

// Handle processes the webhook request
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing Clerk webhook request")

	var event ClerkWebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.logger.Errorw("Failed to decode webhook payload", "error", err)
		render.ChiErr(w, r, err, FailedToDecodeWebhook,
			render.WithStatusCode(http.StatusBadRequest))
		return
	}

	h.logger.Infow("Received webhook event",
		"type", event.Type,
		"object", event.Object,
		"user_id", event.Data.ID)

	// Only process user.created events
	if event.Type != "user.created" {
		h.logger.Warnw("Unsupported event type", "type", event.Type)
		render.ChiErr(w, r, nil, UnsupportedEventType,
			render.WithStatusCode(http.StatusBadRequest))
		return
	}

	// Validate that we have the required data
	if event.Data.ID == "" {
		h.logger.Error("Missing user ID in webhook payload")
		render.ChiErr(w, r, nil, InvalidWebhookPayload,
			render.WithStatusCode(http.StatusBadRequest))
		return
	}

	// Create or get existing user
	user, err := h.userDAO.CreateUserFromClerk(r.Context(), event.Data)
	if err != nil {
		h.logger.Errorw("Failed to create user from Clerk data", "error", err, "clerk_id", event.Data.ID)
		render.ChiErr(w, r, err, FailedToCreateUser,
			render.WithStatusCode(http.StatusInternalServerError))
		return
	}

	h.logger.Infow("Successfully processed user.created webhook",
		"clerk_id", event.Data.ID,
		"user_id", user.ID)

	// Return success response
	response := map[string]interface{}{
		"message":  "User created successfully",
		"user_id":  user.ID,
		"clerk_id": event.Data.ID,
	}

	render.ChiJSON(w, r, response)
}
