package conf

import "github.com/caarlos0/env/v11"

// Auth configures OAuth via goth. A provider is enabled only when its client ID and secret are set.
type Auth struct {
	// BaseURL is the externally visible origin of the API, used to build OAuth callback URLs.
	BaseURL string `env:"AUTH_BASE_URL" envDefault:"http://localhost:8000"`
	// FrontendURL is where users are sent after login/logout.
	FrontendURL   string `env:"AUTH_FRONTEND_URL" envDefault:"http://localhost:8000"`
	SessionSecret string `env:"AUTH_SESSION_SECRET"`
	SecureCookies bool   `env:"AUTH_SECURE_COOKIES" envDefault:"false"`

	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`

	AzureClientID     string `env:"AZURE_CLIENT_ID"`
	AzureClientSecret string `env:"AZURE_CLIENT_SECRET"`
	// AzureTenant is a tenant ID, or one of common/organizations/consumers.
	AzureTenant string `env:"AZURE_TENANT" envDefault:"common"`
}

// Parse assumes godotenv has already been loaded by DB.Parse.
func (a *Auth) Parse() error {
	return env.Parse(a)
}
