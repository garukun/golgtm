package config_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/garukun/golgtm/pkg/lgtm/config"
)

func TestConfigFromEnv(t *testing.T) {
	tests := []struct {
		err  bool
		env  map[string]string
		conf *config.Config
	}{
		// Default values
		{
			err: false,
			env: map[string]string{
				"LGTM_GITHUB_SECRET":     "matrix",
				"LGTM_GITHUB_AUTH_TOKEN": "keymaker",
				"LGTM_GITHUB_OWNER":      "garukun",
				"LGTM_GITHUB_REPO":       "golgtm",
			},
			conf: &config.Config{
				Github: config.ConfigGithub{
					Secret:    "matrix",
					AuthToken: "keymaker",
					Owner:     "garukun",
					Repo:      "golgtm",
				},
				Workflow: config.ConfigWorkflow{
					Context: config.ConfigWorkflowContext{
						Name:        "LGTM Code Review",
						Description: "LGTM Code Review workflow.",
						URL:         "https://github.com/garukun/golgtm",
					},

					InReview: config.ConfigWorkflowInReview{
						Label: "Needs Review",
						Trigger: config.NewTrigger(map[string]int{
							"ptal":          1,
							"please review": 1,
							":-1:":          1,
						}),
					},

					Approved: config.ConfigWorkflowApproved{
						Label: "Ready",
						Trigger: config.NewTrigger(map[string]int{
							"lgtm": 1,
							":+1:": 1,
						}),
					},
				},
			},
		},
		// Custom values
		{
			err: false,
			env: map[string]string{
				"LGTM_GITHUB_SECRET":             "matrix",
				"LGTM_GITHUB_AUTH_TOKEN":         "keymaker",
				"LGTM_GITHUB_OWNER":              "garukun",
				"LGTM_GITHUB_REPO":               "golgtm",
				"LGTM_WORKFLOW_CONTEXT_NAME":     "custom context",
				"LGTM_WORKFLOW_INREVIEW_LABEL":   "custom label",
				"LGTM_WORKFLOW_INREVIEW_TRIGGER": "trigger1:1,trigger 2:2",
			},
			conf: &config.Config{
				Github: config.ConfigGithub{
					Secret:    "matrix",
					AuthToken: "keymaker",
					Owner:     "garukun",
					Repo:      "golgtm",
				},
				Workflow: config.ConfigWorkflow{
					Context: config.ConfigWorkflowContext{
						Name:        "custom context",
						Description: "LGTM Code Review workflow.",
						URL:         "https://github.com/garukun/golgtm",
					},

					InReview: config.ConfigWorkflowInReview{
						Label: "custom label",
						Trigger: config.NewTrigger(map[string]int{
							"trigger1":  1,
							"trigger 2": 2,
						}),
					},

					Approved: config.ConfigWorkflowApproved{
						Label: "Ready",
						Trigger: config.NewTrigger(map[string]int{
							"lgtm": 1,
							":+1:": 1,
						}),
					},
				},
			},
		},
		// Missing required values
		{
			err: true,
			env: map[string]string{
				"LGTM_GITHUB_SECRET": "matrix",
				"LGTM_GITHUB_OWNER":  "garukun",
				"LGTM_GITHUB_REPO":   "golgtm",
			},
			conf: nil,
		},
	}

	withTestEnv := func(env map[string]string, fn func()) {
		for k, v := range env {
			os.Setenv(k, v)
		}

		fn()

		for k := range env {
			os.Unsetenv(k)
		}
	}

	for i, test := range tests {
		t.Logf("Testing %d...", i)
		withTestEnv(test.env, func() {
			conf, err := config.ConfigFromEnv()

			if test.err && err == nil || !test.err && err != nil {
				t.Errorf("The reutrned error %v did not met with expectation.", err)
			} else if !reflect.DeepEqual(test.conf, conf) {
				t.Logf("\nExpected: %+v\n  Actual: %+v", test.conf, conf)
				t.Error("Expected configuration from environment did not match with actual.")
			}
		})
	}
}
