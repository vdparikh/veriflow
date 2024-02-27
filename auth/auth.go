package auth

import (
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/vdparikh/veriflow/config"
	"golang.org/x/oauth2"
)

// This allows OIDC and auth connection for various providers
// Right now it only allows OIDC

type Authenticator interface {
	GenerateAuthLink(fn, uuid string) (string, error)
	GetConfig() (*oidc.Provider, oauth2.Config, error)
	GetAccessToken(code string) (*oidc.UserInfo, *oauth2.Token, error)
}

func NewAuthenticator(cfg *config.Config) Authenticator {
	switch cfg.Auth.Provider {
	case "oidc":
		return &OIDCProvider{
			ClientID:     cfg.Auth.ClientID,
			ClientSecret: cfg.Auth.ClientSecret,
			Issuer:       cfg.Auth.Issuer,
			CallbackURL:  fmt.Sprintf("%s/%s", cfg.BaseURL, cfg.Auth.CallbackURL),
		}
	default:
		return nil
	}
}
