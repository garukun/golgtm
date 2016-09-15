package pr

import "github.com/google/go-github/github"

type issue struct {
	*github.Issue
}

// LabelsWithout method returns a list of label names without the given label names of the
// issue.
func (i issue) LabelsWithout(labels ...string) []string {
	if len(i.Labels) == 0 {
		return nil
	}

	// Create a temporary label name lookup.
	ls := make(map[string]struct{})
	for _, l := range labels {
		ls[l] = struct{}{}
	}

	result := make([]string, 0, len(i.Labels))
	for _, l := range i.Labels {
		n := *l.Name
		if _, ok := ls[n]; ok {
			continue
		}

		result = append(result, n)
	}

	return result
}
