package main

import (
	"flag"
	"golang.org/x/oauth2"
	"log"
	"os"
	"strconv"
)

var (
	AuthToken                 oauth2.TokenSource
	Owner, Repo               string
	PR                        int
	ApprovalTrigger           string
	ApprovalCount             int
	LabelNotReady, LabelReady string
)

var flagPR = flag.Int("pr", 0, "PR number to check for LGTM.")

func init() {
	t := fatalAssignEnv("GITHUB_AUTH_TOKEN")
	AuthToken = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t})

	Owner = fatalAssignEnv("LGTM_GITHUB_OWNER")
	Repo = fatalAssignEnv("LGTM_GITHUB_REPO")

	pr := os.Getenv("LGTM_GITHUB_PR")
	if len(pr) == 0 {
		pr = "0"
	}

	if v, err := strconv.Atoi(pr); err != nil {
		log.Fatal(err)
	} else {
		PR = v
	}

	ApprovalTrigger = fatalAssignEnv("LGTM_APPROVAL_TRIGGER")
	count := fatalAssignEnv("LGTM_APPROVAL_COUNT")

	if v, err := strconv.Atoi(count); err != nil {
		log.Fatal(err)
	} else {
		ApprovalCount = v
	}

	LabelNotReady = fatalAssignEnv("LGTM_GITHUB_LABEL_NOT_READY")
	LabelReady = fatalAssignEnv("LGTM_GITHUB_LABEL_READY")
}

func fatalAssignEnv(env string) string {
	v := os.Getenv(env)
	if len(v) == 0 {
		log.Fatalf("Missing %s\n", env)
	}

	return v
}
