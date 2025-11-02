
	package oauthhelper

import (
	"context"
	"encoding/json"
	"time"

	apperr "github.com/zuhrulumam/pm-tool/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	config *oauth2.Config
}

func NewGoogleProvider(dep Dependencies) *GoogleProvider {
	scopes := dep.Scopes
	if len(scopes) == 0 {
		// Default Google scopes
		scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		}
	}

	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     dep.ClientID,
			ClientSecret: dep.ClientSecret,
			RedirectURL:  dep.RedirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *GoogleProvider) GetLoginURL(ctx context.Context, state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *GoogleProvider) GetUserInfo(ctx context.Context, code, state string) (UserInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return UserInfo{}, apperr.Propagate(err, apperr.CodeExternalAPI, "google token exchange failed", 500)
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return UserInfo{}, apperr.Propagate(err, apperr.CodeExternalAPI, "failed to get google user info", 500)
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return UserInfo{}, apperr.Propagate(err, apperr.CodeExternalAPI, "failed to decode google user info", 500)
	}

	return UserInfo{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Picture:  googleUser.Picture,
		Provider: ProviderGoogle,
	}, nil
}

	