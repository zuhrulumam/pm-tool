package oauthhelper

import (
	"context"
	"fmt"

	apperr "github.com/zuhrulumam/pm-tool/pkg/errors"
)

const (
	ProviderGoogle = "google"
	ProviderGitHub = "github"
)

type Dependencies struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	ProviderType string
}

// Provider defines the interface that all OAuth providers must implement
type Provider interface {
	GetLoginURL(ctx context.Context, state string) string
	GetUserInfo(ctx context.Context, code, state string) (UserInfo, error)
}

type Oauth struct {
	provider Provider
}

// NewOauthProvider creates a new OAuth instance with the specified provider
func NewOauthProvider(dep Dependencies) (Oauth, error) {
	var provider Provider

	switch dep.ProviderType {
	case ProviderGoogle:
		provider = NewGoogleProvider(dep)
	case ProviderGitHub:
		provider = NewGitHubProvider(dep)
	default:
		return Oauth{}, apperr.ValidationError(fmt.Sprintf("unsupported provider type: %s", dep.ProviderType))
	}

	return Oauth{
		provider: provider,
	}, nil
}

// NewGoogleOauth creates a new OAuth instance with Google provider (backward compatible)
func NewGoogleOauth(dep Dependencies) Oauth {
	return Oauth{
		provider: NewGoogleProvider(dep),
	}
}

func (o *Oauth) GetUserInfo(ctx context.Context, req GetUserInfoReq) (UserInfo, error) {
	if req.State != "state-token" {
		return UserInfo{}, apperr.ValidationError("invalid state")
	}

	return o.provider.GetUserInfo(ctx, req.Code, req.State)
}

func (o *Oauth) GetLoginUrl(ctx context.Context, req GetUserInfoReq) string {
	return o.provider.GetLoginURL(ctx, req.State)
}

