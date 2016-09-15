package lgtm

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/garukun/golgtm/pkg/http/httpadapter"
	"github.com/garukun/golgtm/pkg/lgtm/config"
	"github.com/garukun/golgtm/pkg/lgtm/internal/adapters"
	"github.com/garukun/golgtm/pkg/lgtm/internal/pr"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Github Event types; see https://developer.github.com/webhooks/#events.
const (
	issueCommentEvent = "issue_comment"
	pullRequestEvent  = "pull_request"
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

	Config config.Config
}

func (l *LGTM) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	l.h.ServeHTTP(resp, req)
}

func New(c *http.Client, conf *config.Config) *LGTM {
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, c)
	oc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Github.AuthToken}))

	g := github.NewClient(oc)
	confCopy := *conf

	u := &pr.Updater{
		Logger: log.New(os.Stdout, "updater", log.LstdFlags),
		G:      g,
		Config: &confCopy,
	}
	u.Start()

	l := &LGTM{
		G:      g,
		Config: confCopy,
	}
	h := adapters.Adapt(
		http.NotFoundHandler(),

		&adapters.Validator{Secret: []byte(conf.Github.Secret)},
		&adapters.EventRouter{
			Events: map[string]httpadapter.Adapter{
				issueCommentEvent: &adapters.IssueComment{
					Updater: u,
					Config:  &confCopy,
				},
				pullRequestEvent: &adapters.PullRequest{
					Updater: u,
					Config:  &confCopy,
					G:       g,
				},
			},
		},
	)

	l.h = h
	return l
}

func ConfigFromEnv() (*config.Config, error) {
	return config.NewFromEnv()
}
