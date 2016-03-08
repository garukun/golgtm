package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
)

const HTTP_HEADER_GOLGTM_HOOK = "X-Golgtm-Hook"

var (
	hookSecret = flag.String("hook_secret", "", "Webhook secret.")
	hookPort   = flag.Int("hook_port", -1, "Hook service port, -1 if service is not to be started.")
)

func HandleHook(resp http.ResponseWriter, req *http.Request, issues *github.IssuesService) {
	if req.Method != "POST" || req.Header.Get("X-GitHub-Event") != "issue_comment" {
		resp.Header().Set(HTTP_HEADER_GOLGTM_HOOK, "not isscmt")
		resp.WriteHeader(http.StatusNoContent)
		return
	}

	signature := req.Header.Get("X-Hub-Signature")
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print("Read hook payload: ", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	} else if err := validateHookPayload(payload, signature[5:]); err != nil {
		log.Print(err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	event := &github.IssueCommentEvent{}

	if err := json.Unmarshal(payload, event); err != nil {
		log.Print("Unmarshal issue comment: ", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	if event.Issue.PullRequestLinks == nil || event.Comment.Body == nil {
		// Ignore non-pull requests or no comment.
		resp.Header().Set(HTTP_HEADER_GOLGTM_HOOK, "not pr")
		resp.WriteHeader(http.StatusNoContent)
		return
	}

	if !strings.HasPrefix(strings.ToLower(*event.Comment.Body), "lgtm") {
		resp.Header().Set(HTTP_HEADER_GOLGTM_HOOK, "not lgtm")
		resp.WriteHeader(http.StatusNoContent)
		return
	}

	labels := make([]string, 0, len(event.Issue.Labels))
	for _, label := range event.Issue.Labels {
		name := *label.Name

		if name == LabelReady {
			resp.Header().Set(HTTP_HEADER_GOLGTM_HOOK, "has ready")
			resp.WriteHeader(http.StatusNoContent)
			return
		} else if name != LabelNotReady {
			labels = append(labels, name)
		}
	}

	owner := event.Repo.Owner.Login
	repo := event.Repo.Name
	pr := event.Issue.Number
	labels = append(labels, LabelReady)

	if owner == nil || repo == nil || pr == nil {
		resp.Header().Set(HTTP_HEADER_GOLGTM_HOOK, "no orp")
		resp.WriteHeader(http.StatusNoContent)
		return
	}

	if _, _, err := issues.ReplaceLabelsForIssue(*owner, *repo, *pr, labels); err != nil {
		log.Printf("Cannot replace labels: %v, %v", labels, err)
		resp.WriteHeader(http.StatusBadRequest)
	} else {
		resp.Write([]byte("Done!"))
	}
}

func validateHookPayload(payload []byte, signature string) error {
	mac := hmac.New(sha1.New, []byte(*hookSecret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))
	if sig != signature {
		return fmt.Errorf("Invalid signature from Github: %s", signature)
	}

	return nil
}
