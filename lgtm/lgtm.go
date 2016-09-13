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
// The LGTM workflow can be roughly summarized as:
// 1) Create a new pull request;
// 2) Comments on a pull request;
// 3) Label changes to the pull request.
type LGTM struct {
	http.Handler

	G *github.Client
}

func New(c *http.Client, authToken string) *LGTM {
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, c)
	oc := oauth2.NewClient(ctx, authToken)

	g := github.NewClient(oc)
	h := adapters.Adapt(
		http.NotFoundHandler(),
		&adapters.Validator{},
		&adapters.EventRouter{
			Events: map[string]httpadapter.Adapter{},
		},
	)

	return &LGTM{Handler: h, G: g}
}
