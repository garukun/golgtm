package lgtm

import (
	"context"
	"net/http"

	"github.com/garukun/golgtm/http/httpadapter"
	"github.com/garukun/golgtm/lgtm/internal/adapters"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// LGTM implements an http.Handler interface and handles incoming GitHub webhook requests to process
// against a common code review process called, LGTM, a.k.a., "Looks good to me!".
//
// The LGTM workflow can be roughly summarized as the management of two states:
// 1) Needs more review;
// 2) Approved/Ready to be merged.
//
// Regardless of the state of the PR, LGTM will manage the lifecycle of the above two states and
// provide relevant webhook context that can be used to gate from the PR being merged.
type LGTM struct {
	h http.Handler

	G *github.Client

	Config Config
}

func (l *LGTM) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	l.h.ServeHTTP(resp, req)
}

func New(c *http.Client, conf *Config) *LGTM {
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, c)
	oc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Github.AuthToken}))

	g := github.NewClient(oc)
	h := adapters.Adapt(
		http.NotFoundHandler(),

		&adapters.Validator{Secret: []byte(conf.Github.Secret)},
		&adapters.EventRouter{
			Events: map[string]httpadapter.Adapter{},
		},
	)

	return &LGTM{
		h:      h,
		G:      g,
		Config: *conf,
	}
}
