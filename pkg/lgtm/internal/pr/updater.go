package pr

import (
	"fmt"
	"log"
	"sync"

	"github.com/garukun/golgtm/pkg/lgtm/config"
	"github.com/google/go-github/github"
)

type Updater struct {
	*log.Logger

	G      *github.Client
	Config *config.Config

	startOnce sync.Once
	updatesCh chan Update
}

func (u *Updater) Updates() chan<- Update {
	return u.updatesCh
}

func (u *Updater) Start() {
	const updateBuffer = 100

	u.startOnce.Do(func() {
		u.updatesCh = make(chan Update, updateBuffer)
	})

	go func(updatesCh <-chan Update) {
		gconf := u.Config.Github
		w := u.Config.Workflow

		var label, status string

		for up := range updatesCh {
			switch up.State {
			case InReview:
				label = w.InReview.Label
				status = "pending"
			case Approved:
				label = w.Approved.Label
				status = "success"
			}

			u.Printf("appending label %s and status %s", label, status)
			if up.Issue != nil {
				labels := issue{up.Issue}.LabelsWithout(w.InReview.Label, w.Approved.Label)
				labels = append(labels, label)

				if _, _, err := u.G.Issues.ReplaceLabelsForIssue(gconf.Owner, gconf.Repo, up.Number, labels); err != nil {
					u.Print(fmt.Errorf("cannot replace labels %v, %v", labels, err))
				}
			}

			if up.PullRequest != nil {
				ref := *up.PullRequest.Head.SHA
				rs := &github.RepoStatus{
					State:       &status,
					TargetURL:   &w.Context.URL,
					Context:     &w.Context.Name,
					Description: &w.Context.Description,
				}

				if _, _, err := u.G.Repositories.CreateStatus(gconf.Owner, gconf.Repo, ref, rs); err != nil {
					u.Print(fmt.Errorf("cannot create pending status, %s, %v", ref, err))
				}
			}
		}

		u.Print("No more updates, done!")
	}(u.updatesCh)
}
