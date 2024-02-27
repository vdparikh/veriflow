package auth

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	CallbackURL  string
}

func (o *OIDCProvider) GetConfig() (*oidc.Provider, oauth2.Config, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, o.Issuer)
	if err != nil {
		return nil, oauth2.Config{}, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		RedirectURL:  o.CallbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return provider, oauth2Config, nil
}

func (o *OIDCProvider) GenerateAuthLink(fn string, uuid string) (string, error) {
	stateString := fmt.Sprintf("%s=%s", fn, uuid)
	state := base64.StdEncoding.EncodeToString([]byte(stateString))

	_, oauth2Config, err := o.GetConfig()
	return oauth2Config.AuthCodeURL(state), err
}

func (o *OIDCProvider) GetAccessToken(code string) (*oidc.UserInfo, *oauth2.Token, error) {
	ctx := context.Background()
	provider, config, _ := o.GetConfig()

	oauth2Token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
	return userInfo, oauth2Token, err
}
