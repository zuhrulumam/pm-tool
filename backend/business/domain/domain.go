package domain

import (
	"github.com/zuhrulumam/pm-tool/config"
	"github.com/zuhrulumam/pm-tool/infra/redis"
	"github.com/zuhrulumam/pm-tool/pkg/httpclient"
	oauthhelper "github.com/zuhrulumam/pm-tool/pkg/oauth"
	"go.opentelemetry.io/otel/trace"
	"github.com/zuhrulumam/pm-tool/pkg/db"
)

// Domains holds all domain instances
// Schema: public.
type Domains struct {
	User UserDomainItf
	Project ProjectDomainItf
	Note NoteDomainItf
}

// DomainDependencies contains dependencies needed to initialize domains
type DomainDependencies struct {
	DB           *db.DB  // Can be single DB or use Leader/Follower pattern
	Redis        *redis.Client
	HTTP         *httpclient.Client
	Tracer       trace.Tracer
	Oauth        oauthhelper.Oauth
	SchemaPrefix string
	Config *config.Config
}

// NewDomains creates and initializes all domains with dependencies
// All domains are schema-aware and will use: public.
func NewDomains(deps DomainDependencies) *Domains {
	return &Domains{
		User: NewUserDomain(
			deps.DB,
			deps.Config,
			deps.Redis,
			deps.HTTP,
			deps.Tracer,
			deps.SchemaPrefix,
		),
		Project: NewProjectDomain(
			deps.DB,
			deps.Config,
			deps.Redis,
			deps.HTTP,
			deps.Tracer,
			deps.SchemaPrefix,
		),
		Note: NewNoteDomain(
			deps.DB,
			deps.Config,
			deps.Redis,
			deps.HTTP,
			deps.Tracer,
			deps.SchemaPrefix,
		),
	}
}
