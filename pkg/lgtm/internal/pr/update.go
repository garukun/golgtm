package pr

import "github.com/google/go-github/github"

type Update struct {
	Number int // Issue number, aka. PR number
	State  State

	// The fields belong should be set based on the context of the Github Event API, e.g., certain
	// events does not directly pass an issue. It's up to the Updater's implementation on how to
	// interpret missing fields; clients should not perform additional API call to fill in the missing
	// fields.

	Issue       *github.Issue // Issue must refer to a PR
	PullRequest *github.PullRequest
}

type State uint8

const (
	InReview State = iota
	Approved
)
