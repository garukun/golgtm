package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/garukun/golgtm/pkg/lgtm/config"
	"github.com/garukun/golgtm/pkg/lgtm/internal/pr"
	"github.com/google/go-github/github"
)

// IssueComment handles when a GitHub issue comment event is fired.
type IssueComment struct {
	*pr.Updater

	Config *config.Config
}

func (c *IssueComment) Adapt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		event := &github.IssueCommentEvent{}

		if err := json.NewDecoder(req.Body).Decode(event); err != nil {
			log.Print("unmarshal: ", err)
			resp.Header().Set(ResponseHeader, "issue comment fmt")
			resp.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := c.validate(event); err != nil {
			log.Printf("validate: %v", err)
			resp.Header().Set(ResponseHeader, err.Error())
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		update, err := c.newUpdate(event)
		if err != nil {
			log.Printf("issue comment no update: %v", err)
			resp.Header().Set(ResponseHeader, err.Error())
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		// Enqueue the pending updates.
		// TODO(@garukun): If we want to report error based on underlying API errors, we'll need to pass
		// ResponseWriter through the channel.
		c.Updates() <- *update

		log.Printf("Updated LGTM for %s/%s#%d!", c.Config.Github.Owner, c.Config.Github.Repo, *event.Issue.Number)
		resp.Write([]byte("Done!"))

		// Swallow downstream handlers?
	})
}

func (c *IssueComment) validate(e *github.IssueCommentEvent) error {
	// TODO(@garukun): Should e.Repo.Owner.Login, e.Repo.Name, e.Issue.Number against config

	if e.Issue.PullRequestLinks == nil {
		// Ignore non-pull requests
		return errors.New("not pr")
	}

	if e.Comment.Body == nil || len(strings.TrimSpace(*e.Comment.Body)) == 0 {
		// Ignore no comments
		return fmt.Errorf("no comment, %v", e.Comment.Body)
	}

	return nil
}

func (c *IssueComment) newUpdate(e *github.IssueCommentEvent) (*pr.Update, error) {
	update, err := c.checkTriggers(*e.Comment.Body)
	if err != nil {
		return nil, err
	}

	var label string
	switch update.State {
	case pr.InReview:
		label = c.Config.Workflow.InReview.Label
	case pr.Approved:
		label = c.Config.Workflow.Approved.Label
	}

	if !c.shouldUpdateLabels(e.Issue.Labels, label) {
		return nil, fmt.Errorf("already labeled: %s", label)
	}

	update.Issue = e.Issue
	update.Number = *e.Issue.Number
	return update, nil
}

func (c *IssueComment) checkTriggers(comment string) (*pr.Update, error) {
	comment = strings.ToLower(strings.TrimSpace(comment))
	for t := range c.Config.Workflow.Approved.Trigger {
		if strings.HasPrefix(comment, t) && strings.HasSuffix(comment, t) {
			return &pr.Update{State: pr.Approved}, nil
		}
	}

	for t := range c.Config.Workflow.InReview.Trigger {
		if strings.HasPrefix(comment, t) && strings.HasSuffix(comment, t) {
			return &pr.Update{State: pr.InReview}, nil
		}
	}

	log.Printf("no lgtm triggers: approved:%v, in-review:%v", c.Config.Workflow.Approved.Trigger, c.Config.Workflow.InReview.Trigger)
	return nil, errors.New("no lgtm triggers")
}

func (c *IssueComment) shouldUpdateLabels(labels []github.Label, name string) bool {
	return !githubLabels(labels).Contains(name)
}

type githubLabels []github.Label

// Contains method return whether a given label name is in a list of GitHub labels.
func (l githubLabels) Contains(name string) bool {
	for _, label := range l {
		if *label.Name == name {
			return true
		}
	}

	return false
}
