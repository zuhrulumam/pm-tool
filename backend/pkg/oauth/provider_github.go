package oauthhelper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apperr "github.com/zuhrulumam/pm-tool/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubProvider struct {
	config *oauth2.Config
}

func NewGitHubProvider(dep Dependencies) *GitHubProvider {
	scopes := dep.Scopes
	if len(scopes) == 0 {
		// Default GitHub scopes
		scopes = []string{"user:email"}
	}

	return &GitHubProvider{
		config: &oauth2.Config{
			ClientID:     dep.ClientID,
			ClientSecret: dep.ClientSecret,
			RedirectURL:  dep.RedirectURL,
			Scopes:       scopes,
			Endpoint:     github.Endpoint,
		},
	}
}

func (gh *GitHubProvider) GetLoginURL(ctx context.Context, state string) string {
	return gh.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (gh *GitHubProvider) GetUserInfo(ctx context.Context, code, state string) (UserInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	token, err := gh.config.Exchange(ctx, code)
	if err != nil {
		return UserInfo{}, apperr.Propagate(err, apperr.CodeExternalAPI, "github token exchange failed", 500)
	}

	client := gh.config.Client(ctx, token)

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return UserInfo{}, apperr.Propagate(err, apperr.CodeExternalAPI, "failed to get github user info", 500)
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return UserInfo{}, apperr.Propagate(err, apperr.CodeExternalAPI, "failed to decode github user info", 500)
	}

	// GitHub doesn't always return email in /user endpoint
	// If email is empty, fetch from /user/emails
	email := githubUser.Email
	if email == "" {
		email, err = gh.getUserEmail(ctx, client)
		if err != nil {
			return UserInfo{}, err
		}
	}

	// GitHub uses int64 for ID, convert to string
	userID := fmt.Sprintf("%d", githubUser.ID)

	name := githubUser.Name
	if name == "" {
		name = githubUser.Login // Fallback to login if name is not set
	}

	return UserInfo{
		ID:       userID,
		Email:    email,
		Name:     name,
		Picture:  githubUser.AvatarURL,
		Provider: ProviderGitHub,
	}, nil
}

func (gh *GitHubProvider) getUserEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", apperr.Propagate(err, apperr.CodeExternalAPI, "failed to get github user emails", 500)
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", apperr.Propagate(err, apperr.CodeExternalAPI, "failed to decode github emails", 500)
	}

	// Find primary verified email
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	// Fallback to first verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", apperr.ValidationError("no verified email found for github user")
}

func (gh *GitHubProvider) VerifyIDToken(ctx context.Context, idToken string) (UserInfo, error) {
	return UserInfo{}, nil
}
