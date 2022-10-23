package server

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/handler"
	"github.com/teamhanko/hanko/backend/persistence"
	hankoMiddleware "github.com/teamhanko/hanko/backend/server/middleware"
	"github.com/teamhanko/hanko/backend/session"
)

func NewPrivateRouter(cfg *config.Config, persister persistence.Persister) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.HTTPErrorHandler = dto.NewHTTPErrorHandler(dto.HTTPErrorHandlerConfig{Debug: false, Logger: e.Logger})
	e.Use(middleware.RequestID())
	e.Use(hankoMiddleware.GetLoggerMiddleware())

	jwkManager, err := jwk.NewDefaultManager(cfg.Secrets.Keys, persister.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, cfg.Session, persister)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}

	e.Validator = dto.NewCustomValidator()

	healthHandler := handler.NewHealthHandler()

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	userHandler := handler.NewUserHandlerAdmin(persister)

	user := e.Group("/users")
	user.DELETE("/:id", userHandler.Delete, hankoMiddleware.Session(sessionManager))
	user.PATCH("/:id", userHandler.Patch, hankoMiddleware.Session(sessionManager))
	user.GET("", userHandler.List, hankoMiddleware.Session(sessionManager))

	return e
}
