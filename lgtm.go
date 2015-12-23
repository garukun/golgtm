package main

import (
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

type LGTM struct {
	G *github.Client

	// Cached objects.
	comments []github.IssueComment
	labels   []github.Label
}

func (l *LGTM) ReadComments() []github.IssueComment {
	l.comments = fatal(l.G.Issues.ListComments(Owner, Repo, PR, nil)).([]github.IssueComment)
	return l.comments
}

func (l *LGTM) ReadLabels() []github.Label {
	l.labels = fatal(l.G.Issues.ListLabelsByIssue(Owner, Repo, PR, nil)).([]github.Label)
	return l.labels
}

// IsApproved returns whether there are enough comments on the PR to warrant the PR approval.
func (l *LGTM) IsApproved() bool {
	if len(l.comments) == 0 {
		l.ReadComments()
	}

	count := ApprovalCount
	for _, c := range l.comments {
		if *c.Body == ApprovalTrigger {
			count--
		}

		if count == 0 {
			return true
		}
	}

	return false
}

// Unapprove checks the labels and make sure that the READY label is not present, and NOT_READY label
// is created.
func (l *LGTM) Unapprove() {
	nr, r := l.labelCheck()

	if !nr {
		l.G.Issues.AddLabelsToIssue(Owner, Repo, PR, []string{LabelNotReady})
	}

	if r {
		l.G.Issues.RemoveLabelForIssue(Owner, Repo, PR, LabelReady)
	}
}

// Approve checks the labels and make sure that the READY label is present, and NOT_READY label
// is rmeoved.
func (l *LGTM) Approve() {
	nr, r := l.labelCheck()

	if nr {
		l.G.Issues.AddLabelsToIssue(Owner, Repo, PR, []string{LabelReady})
	}

	if !r {
		l.G.Issues.RemoveLabelForIssue(Owner, Repo, PR, LabelNotReady)
	}
}

func (l *LGTM) labelCheck() (hasNotReady bool, hasReady bool) {
	if len(l.labels) == 0 {
		l.ReadLabels()
	}

	for _, label := range l.labels {
		if n := *label.Name; n == LabelNotReady {
			hasNotReady = true
		} else if n == LabelReady {
			hasReady = true
		}

		if hasReady && hasNotReady {
			return true, true
		}
	}

	return
}

func NewLGTM(certClient *http.Client) *LGTM {
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, certClient)
	c := oauth2.NewClient(ctx, AuthToken)
	g := github.NewClient(c)

	return &LGTM{G: g}
}

func fatal(v interface{}, resp *github.Response, err error) interface{} {

	if err != nil && resp != nil {
		log.Fatal(resp, err)
	} else if err != nil {
		log.Fatal(err)
	}

	return v
}
