package server

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/handler"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	hankoMiddleware "github.com/teamhanko/hanko/backend/server/middleware"
	"github.com/teamhanko/hanko/backend/server/ws"
	"github.com/teamhanko/hanko/backend/session"
)

func NewPublicRouter(cfg *config.Config, persister persistence.Persister) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.HTTPErrorHandler = dto.NewHTTPErrorHandler(dto.HTTPErrorHandlerConfig{Debug: false, Logger: e.Logger})
	e.Use(middleware.RequestID())
	e.Use(hankoMiddleware.GetLoggerMiddleware())

	if cfg.Server.Public.Cors.Enabled {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.Server.Public.Cors.AllowOrigins,
			AllowMethods:     cfg.Server.Public.Cors.AllowMethods,
			AllowHeaders:     cfg.Server.Public.Cors.AllowHeaders,
			ExposeHeaders:    cfg.Server.Public.Cors.ExposeHeaders,
			AllowCredentials: cfg.Server.Public.Cors.AllowCredentials,
			MaxAge:           cfg.Server.Public.Cors.MaxAge,
		}))
	}

	e.Validator = dto.NewCustomValidator()

	jwkManager, err := jwk.NewDefaultManager(cfg.Secrets.Keys, persister.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, cfg.Session, persister)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}

	mailer, err := mail.NewMailer(cfg.Passcode.Smtp)
	if err != nil {
		panic(fmt.Errorf("failed to create mailer: %w", err))
	}

	if cfg.Password.Enabled {
		passwordHandler := handler.NewPasswordHandler(persister, sessionManager, cfg)

		password := e.Group("/password")
		password.PUT("", passwordHandler.Set, hankoMiddleware.Session(sessionManager))
		password.POST("/login", passwordHandler.Login)
	}

	userHandler := handler.NewUserHandler(cfg, persister, sessionManager)

	e.GET("/me", userHandler.Me, hankoMiddleware.Session(sessionManager))
	e.POST("/login/guest", userHandler.InitiateLoginAsGuest, hankoMiddleware.Session(sessionManager))

	user := e.Group("/users")
	user.POST("", userHandler.Create)
	user.GET("/:id", userHandler.Get, hankoMiddleware.Session(sessionManager))
	user.POST("/logout", userHandler.Logout, hankoMiddleware.Session(sessionManager))
	user.POST("/logout-guest", userHandler.LogoutAsGuest, hankoMiddleware.Session(sessionManager))
	user.GET("/shares/overview", userHandler.GetUserGuestRelationsOverview, hankoMiddleware.Session(sessionManager))
	user.GET("/shares/guest", userHandler.GetUserGuestRelationsAsGuest, hankoMiddleware.Session(sessionManager))
	user.GET("/shares/parent", userHandler.GetUserGuestRelationsAsAccountHolder, hankoMiddleware.Session(sessionManager))
	user.DELETE("/shares/:id", userHandler.RemoveAccessToRelation, hankoMiddleware.Session(sessionManager))

	e.POST("/user", userHandler.GetUserIdByEmail)

	healthHandler := handler.NewHealthHandler()
	webauthnHandler, err := handler.NewWebauthnHandler(cfg, persister, sessionManager)
	if err != nil {
		panic(fmt.Errorf("failed to create public webauthn handler: %w", err))
	}
	passcodeHandler, err := handler.NewPasscodeHandler(cfg, persister, sessionManager, mailer)
	if err != nil {
		panic(fmt.Errorf("failed to create public passcode handler: %w", err))
	}
	accountSharingHandler, err := handler.NewAccountSharingHandler(cfg, persister, sessionManager, mailer)
	if err != nil {
		panic(fmt.Errorf("failed to create public account sharing handler: %w", err))
	}
	websocketHandler, err := ws.NewWebsocketHandler(cfg, persister, sessionManager, accountSharingHandler)
	if err != nil {
		panic(fmt.Errorf("failed to create websocker handler: %w", err))
	}

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	wellKnownHandler, err := handler.NewWellKnownHandler(*cfg, jwkManager)
	if err != nil {
		panic(fmt.Errorf("failed to create well-known handler: %w", err))
	}
	wellKnown := e.Group("/.well-known")
	wellKnown.GET("/jwks.json", wellKnownHandler.GetPublicKeys)
	wellKnown.GET("/config", wellKnownHandler.GetConfig)

	webauthn := e.Group("/webauthn")
	webauthnRegistration := webauthn.Group("/registration", hankoMiddleware.Session(sessionManager))
	webauthnRegistration.POST("/initialize", webauthnHandler.BeginRegistration)
	webauthnRegistration.POST("/finalize", webauthnHandler.FinishRegistration)

	webauthnLogin := webauthn.Group("/login")
	webauthnLogin.POST("/initialize", webauthnHandler.BeginAuthentication)
	webauthnLogin.POST("/finalize", webauthnHandler.FinishAuthentication)

	access := e.Group("/access")
	share := access.Group("/share", hankoMiddleware.Session(sessionManager))
	share.POST("/initialize", accountSharingHandler.BeginShare)
	share.POST("/begin-create-account-with-grant", accountSharingHandler.BeginCreateAccountWithGrant)
	share.POST("/finish-create-account-with-grant", accountSharingHandler.FinishCreateAccountWithGrant)

	passcode := e.Group("/passcode")
	passcodeLogin := passcode.Group("/login")
	passcodeLogin.POST("/initialize", passcodeHandler.Init)
	passcodeLogin.POST("/finalize", passcodeHandler.Finish)

	signatureFakerHandler := handler.NewSignatureFakerHandler(persister)
	signature := e.Group("/sign")
	signature.POST("", signatureFakerHandler.SignChallengeAsUser)

	e.GET("/ws/:id", websocketHandler.WsPage, hankoMiddleware.Session(sessionManager))

	adminHandler := handler.NewUserHandlerAdmin(persister)
	admin := e.Group("/admin")
	admin.GET("/users", adminHandler.List, hankoMiddleware.Session(sessionManager))
	admin.GET("/grants/:id", adminHandler.GetGrantsForUser, hankoMiddleware.Session(sessionManager))
	admin.POST("/login-audit", adminHandler.GetLoginAuditRecordsForUser, hankoMiddleware.Session(sessionManager))
	admin.PUT("/users/active/:id", adminHandler.ToggleIsActiveForUser, hankoMiddleware.Session(sessionManager))
	admin.DELETE("/grants/:id", adminHandler.DeactivateGrantsForUser, hankoMiddleware.Session(sessionManager))

	postHandler := handler.NewPostHandler(persister)
	posts := e.Group("/posts")
	posts.GET("", postHandler.GetPosts, hankoMiddleware.Session(sessionManager))
	posts.POST("", postHandler.CreatePost, hankoMiddleware.Session(sessionManager))

	return e
}
