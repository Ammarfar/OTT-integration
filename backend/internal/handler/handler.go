package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ott-integration/backend/internal/config"
	"ott-integration/backend/internal/domain"
	httpdto "ott-integration/backend/internal/dto/http"
	"ott-integration/backend/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.POST("/subscribe", h.subscribe)
		api.POST("/activate", h.activate)
		api.GET("/subscription-status", h.subscriptionStatus)
		api.GET("/providers", h.providers)
	}
}

func (h *Handler) subscribe(c *gin.Context) {
	var req httpdto.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	result, err := h.service.Subscribe(c.Request.Context(), domain.SubscribeCommand{
		UserID:   req.UserID,
		MSISDN:   req.MSISDN,
		Provider: domain.ProviderName(req.Provider),
		Plan:     domain.Plan(req.Plan),
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "success", httpdto.SubscribeResponse{
		SubscriptionRequestID: result.Record.SubscriptionRequestID,
		ActivationCode:        result.Record.ActivationCode,
		ActivationLink:        result.ActivationURL,
		SMSMessage:            result.SMSMessage,
		Status:                string(result.Record.SubscriptionStatus),
	})
}

func (h *Handler) activate(c *gin.Context) {
	var req httpdto.ActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	result, err := h.service.Activate(c.Request.Context(), domain.ActivateCommand{
		ActivationCode: req.ActivationCode,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	record := result.Record
	writeSuccess(c, http.StatusOK, "success", httpdto.ActivateResponse{
		Provider:            string(record.Provider),
		UserID:              record.UserID,
		ActivationStatus:    string(record.ActivationStatus),
		SubscriptionStatus:  string(record.SubscriptionStatus),
		Plan:                string(record.Plan),
		ExternalReferenceID: record.ExternalReferenceID,
		ActivatedAt:         formatTime(record.ActivatedAt),
		Message:             record.Message,
	})
}

func (h *Handler) subscriptionStatus(c *gin.Context) {
	result, err := h.service.SubscriptionStatus(c.Request.Context(), domain.StatusCommand{
		ActivationCode: c.Query("activationCode"),
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	record := result.Record
	writeSuccess(c, http.StatusOK, "success", httpdto.SubscriptionStatusResponse{
		SubscriptionRequestID: record.SubscriptionRequestID,
		UserID:                record.UserID,
		Provider:              string(record.Provider),
		Plan:                  string(record.Plan),
		SubscriptionStatus:    string(record.SubscriptionStatus),
		ActivatedAt:           formatTime(record.ActivatedAt),
		TokenExpiresAt:        formatTime(record.TokenExpiresAt),
		ExternalReferenceID:   record.ExternalReferenceID,
		Message:               record.Message,
	})
}

func (h *Handler) providers(c *gin.Context) {
	items := h.service.Providers()
	response := make([]httpdto.ProviderInfo, 0, len(items))
	for _, item := range items {
		plans := make([]string, 0, len(item.Plans))
		for _, plan := range item.Plans {
			plans = append(plans, string(plan))
		}
		response = append(response, httpdto.ProviderInfo{
			Name:  string(item.Name),
			Plans: plans,
		})
	}

	writeSuccess(c, http.StatusOK, "success", httpdto.ProvidersResponse{Providers: response})
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, httpdto.Response[httpdto.ErrorResponse]{
		Code:    status,
		Message: message,
		Data:    httpdto.ErrorResponse{Code: code, Error: message},
	})
}

func writeSuccess[T any](c *gin.Context, status int, message string, data T) {
	c.JSON(status, httpdto.Response[T]{
		Code:    status,
		Message: message,
		Data:    data,
	})
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(c, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, service.ErrUnsupported):
		writeError(c, http.StatusBadRequest, "unsupported", "unsupported provider or plan")
	case errors.Is(err, service.ErrActivationNotFound):
		writeError(c, http.StatusNotFound, "not_found", "activation code not found")
	case errors.Is(err, context.DeadlineExceeded):
		writeError(c, http.StatusGatewayTimeout, "timeout", "request timed out")
	default:
		writeError(c, http.StatusBadGateway, "provider_error", err.Error())
	}
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func RequestTimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func NewRouter(svc *service.Service, timeout time.Duration) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware("*"))
	router.Use(RequestTimeoutMiddleware(timeout))

	h := New(svc)
	h.RegisterRoutes(router)

	return router
}

func NewRouterFromConfig(svc *service.Service, cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(cfg.CORSAllowedOrigin))
	router.Use(RequestTimeoutMiddleware(cfg.HTTPTimeout))

	h := New(svc)
	h.RegisterRoutes(router)

	return router
}

func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (allowedOrigin == "*" || strings.EqualFold(origin, allowedOrigin)) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
