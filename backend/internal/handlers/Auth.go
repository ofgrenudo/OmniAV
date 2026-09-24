package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/azureadv2"
	"github.com/markbates/goth/providers/google"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"github.com/ofgrenudo/OmniAv/internal/models/conf"
	"gorm.io/gorm"
)

const (
	userSessionName = "omniav_user"
	userIDKey       = "user_id"
	userContextKey  = "user"
)

type AuthHandler struct {
	DB   *gorm.DB
	Conf conf.Auth
}

// NewAuthHandler configures goth's providers and session store (both process-global in goth).
func NewAuthHandler(db *gorm.DB, cfg conf.Auth) (*AuthHandler, error) {
	if cfg.SessionSecret == "" {
		return nil, errors.New("AUTH_SESSION_SECRET must be set")
	}

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.MaxAge(86400 * 7)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = cfg.SecureCookies
	store.Options.SameSite = http.SameSiteLaxMode
	gothic.Store = store

	var providers []goth.Provider
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		providers = append(providers, google.New(
			cfg.GoogleClientID, cfg.GoogleClientSecret,
			cfg.BaseURL+"/api/auth/google/callback", "email", "profile",
		))
	}
	if cfg.AzureClientID != "" && cfg.AzureClientSecret != "" {
		providers = append(providers, azureadv2.New(
			cfg.AzureClientID, cfg.AzureClientSecret,
			cfg.BaseURL+"/api/auth/azure/callback",
			azureadv2.ProviderOptions{Tenant: azureadv2.TenantType(cfg.AzureTenant)},
		))
	}
	if len(providers) == 0 {
		slog.Warn("No OAuth providers configured; set GOOGLE_CLIENT_ID/SECRET and/or AZURE_CLIENT_ID/SECRET")
	}
	goth.UseProviders(providers...)

	return &AuthHandler{DB: db, Conf: cfg}, nil
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.GET("/me", h.RequireAuth(), h.Me)
		auth.POST("/logout", h.Logout)
		auth.GET("/:provider", h.Begin)
		auth.GET("/:provider/callback", h.Callback)
	}
}

// providerName maps the public route name to goth's provider name ("azure" -> "azureadv2").
func providerName(c *gin.Context) string {
	name := c.Param("provider")
	if name == "azure" {
		return "azureadv2"
	}
	return name
}

// withProvider makes gothic pick the provider from the route param; it reads a "provider" query value.
func withProvider(c *gin.Context) *http.Request {
	q := c.Request.URL.Query()
	q.Set("provider", providerName(c))
	req := c.Request.Clone(c.Request.Context())
	req.URL.RawQuery = q.Encode()
	return req
}

func (h *AuthHandler) Begin(c *gin.Context) {
	if _, err := goth.GetProvider(providerName(c)); err != nil {
		notFound(c, "unknown or unconfigured auth provider")
		return
	}
	gothic.BeginAuthHandler(c.Writer, withProvider(c))
}

func (h *AuthHandler) Callback(c *gin.Context) {
	req := withProvider(c)
	gu, err := gothic.CompleteUserAuth(c.Writer, req)
	if err != nil {
		slog.Warn("OAuth callback failed", slog.Any("error", err))
		c.Redirect(http.StatusFound, h.frontendRedirect("login_failed"))
		return
	}

	user := models.User{Provider: gu.Provider, ProviderUserID: gu.UserID}
	if err := h.DB.Where(&user).FirstOrInit(&user).Error; err != nil {
		internalError(c, err)
		return
	}
	user.Email, user.Name, user.AvatarURL = gu.Email, displayName(gu), gu.AvatarURL
	if err := h.DB.Save(&user).Error; err != nil {
		internalError(c, err)
		return
	}

	// Drop goth's provider session and start our own.
	_ = gothic.Logout(c.Writer, req)
	session, _ := gothic.Store.New(c.Request, userSessionName)
	session.Values[userIDKey] = user.ID
	if err := session.Save(c.Request, c.Writer); err != nil {
		internalError(c, err)
		return
	}

	c.Redirect(http.StatusFound, h.Conf.FrontendURL)
}

func displayName(gu goth.User) string {
	if gu.Name != "" {
		return gu.Name
	}
	return gu.Email
}

func (h *AuthHandler) frontendRedirect(errCode string) string {
	return h.Conf.FrontendURL + "?" + url.Values{"auth_error": {errCode}}.Encode()
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session, _ := gothic.Store.Get(c.Request, userSessionName)
	session.Options.MaxAge = -1
	if err := session.Save(c.Request, c.Writer); err != nil {
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, _ := c.Get(userContextKey)
	c.JSON(http.StatusOK, user)
}

// RequireAuth rejects requests without a valid login session and stores the user in the context.
func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := gothic.Store.Get(c.Request, userSessionName)
		id, ok := session.Values[userIDKey].(uint)
		if err != nil || !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		var user models.User
		if err := h.DB.First(&user, id).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.Set(userContextKey, user)
		c.Next()
	}
}
