package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/garukun/golgtm/pkg/lgtm/config"
	"github.com/garukun/golgtm/pkg/lgtm/internal/pr"
	"github.com/google/go-github/github"
)

type PullRequest struct {
	*pr.Updater

	G      *github.Client
	Config *config.Config
}

func (p *PullRequest) Adapt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		event := &github.PullRequestEvent{}

		if err := json.NewDecoder(req.Body).Decode(event); err != nil {
			log.Print("unmarshal: ", err)
			resp.Header().Set(ResponseHeader, "pr fmt")
			resp.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := p.validate(event); err != nil {
			log.Printf("validate: %v", err)
			resp.Header().Set(ResponseHeader, err.Error())
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		update, err := p.newUpdate(event)
		if err != nil {
			log.Printf("pr no update: %v", err)
			resp.Header().Set(ResponseHeader, err.Error())
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		// Enqueue the pending updates.
		// TODO(@garukun): If we want to report error based on underlying API errors, we'll need to pass
		// ResponseWriter through the channel.
		p.Updates() <- *update

		log.Printf("Updated LGTM for %s/%s#%d!", p.Config.Github.Owner, p.Config.Github.Repo, *event.Number)
		resp.Write([]byte("Done!"))

		// Swallow downstream handlers?
	})
}

func (p *PullRequest) validate(e *github.PullRequestEvent) error {
	// TODO(@garukun): Should e.Repo.Owner.Login, e.Repo.Name, e.Issue.Number against config

	action := e.Action
	if action == nil {
		return errors.New("nil pr action")
	}

	switch a := *action; a {
	case "opened", "reopened", "labeled", "unlabeled", "synchronized":
		return nil
	default:
		return fmt.Errorf("invalid action: %s", a)
	}
}

func (p *PullRequest) newUpdate(e *github.PullRequestEvent) (*pr.Update, error) {
	switch *e.Action {
	case "labeled", "unlabeled":
		issue, _, err := p.G.Issues.Get(p.Config.Github.Repo, p.Config.Github.Owner, *e.Number)
		if err != nil {
			return nil, err
		}

		for _, l := range issue.Labels {
			if *l.Name == p.Config.Workflow.Approved.Label {
				return &pr.Update{
					State:       pr.Approved,
					Number:      *e.Number,
					PullRequest: e.PullRequest,
				}, nil
			}
		}
	}

	return &pr.Update{
		State:       pr.InReview,
		Number:      *e.Number,
		PullRequest: e.PullRequest,
	}, nil
}
