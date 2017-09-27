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

const (
	prActionOpened      = "opened"
	prActionReopened    = "reopened"
	prActionLabeled     = "labeled"
	prActionUnlabeled   = "unlabeled"
	prActionSynchronize = "synchronize"
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
	case prActionOpened, prActionReopened, prActionLabeled, prActionUnlabeled, prActionSynchronize:
		return nil
	default:
		return fmt.Errorf("invalid action: %s", a)
	}
}

func (p *PullRequest) newUpdate(e *github.PullRequestEvent) (*pr.Update, error) {
	var updateIssue *github.Issue

	switch *e.Action {
	case prActionSynchronize:
		issue, err := p.getIssue(*e.Number)
		if err != nil {
			return nil, err
		}

		updateIssue = issue

		if !githubLabels(issue.Labels).Contains(p.Config.Workflow.InReview.Label) {
			// Adding comments in a goroutine is a bit racier because from the moment we verified that it
			// doesn't contain InReview comments to when the goroutine gets executed, the labels may have
			// changed.
			go func(p *PullRequest) {
				log.Printf("revert %s/%s#%d review status", p.Config.Github.Owner, p.Config.Github.Repo, *e.Number)

				if err := p.addComment(*e.Number, "Files changed in PR, reverting code review status."); err != nil {
					log.Printf("cannot add comment to %s/%s#%d: %v", p.Config.Github.Owner, p.Config.Github.Repo, *e.Number, err)
				}
			}(p)
		}
	case prActionLabeled, prActionUnlabeled:
		issue, err := p.getIssue(*e.Number)
		if err != nil {
			return nil, err
		}

		if githubLabels(issue.Labels).Contains(p.Config.Workflow.Approved.Label) {
			return &pr.Update{
				State:       pr.Approved,
				Number:      *e.Number,
				PullRequest: e.PullRequest,
			}, nil
		}
	}

	return &pr.Update{
		State:       pr.InReview,
		Number:      *e.Number,
		Issue:       updateIssue,
		PullRequest: e.PullRequest,
	}, nil
}

func (p *PullRequest) getIssue(number int) (*github.Issue, error) {
	issue, _, err := p.G.Issues.Get(p.Config.Github.Owner, p.Config.Github.Repo, number)
	return issue, err
}

func (p *PullRequest) addComment(number int, comment string) error {
	ic := &github.IssueComment{
		Body: &comment,
	}

	_, _, err := p.G.Issues.CreateComment(p.Config.Github.Owner, p.Config.Github.Repo, number, ic)
	return err
}
