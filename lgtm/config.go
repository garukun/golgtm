package lgtm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Github struct {
		Secret    string `envconfig:"secret" required:"true"`
		AuthToken string `envconfig:"auth_token" required:"true"`
		Owner     string `envconfig:"owner" required:"true"`
		Repo      string `envconfig:"repo" required:"true"`
	}

	Workflow struct {
		Context struct {
			Name        string `envconfig:"name" default:"LGTM Code Review"`
			Description string `envconfig:"desc" default:"LGTM Code Review workflow."`
			URL         string `envconfig:"url" default:"https://github.com/garukun/golgtm"`
		}

		InReview struct {
			Label   string  `envconfig:"label" default:"Needs Review"`
			Trigger trigger `envconfig:"trigger" default:"ptal:1,please review:1,:-1::1"`
		}

		Approved struct {
			Label   string  `envconfig:"label" default:"Ready"`
			Trigger trigger `envconfig:"trigger" default:"lgtm:1,:+1::1"`
		}
	}
}

// trigger method implements an envconfig.Decoder interface to provide a custom environment variable
// deserialization format.
//
// Format:
// 	<trigger phrase>:<trigger count>[,<trigger phrase>:<trigger count>]
//
// Trigger phrase can be any string literal except the `,` character; string literal does not need
// to be escaped in any way including the `:` character;
// Trigger number must be a positive integer.
type trigger map[string]int

func (t *trigger) Decode(value string) error {
	if len(value) == 0 {
		return nil
	}

	tmp := make(trigger)
	triggers := strings.Split(value, ",")
	for _, trig := range triggers {
		sep := strings.LastIndex(trig, ":")
		if sep < 0 || sep+1 == len(trig) {
			return fmt.Errorf("Invalid trigger format %s around %s", value, trig)
		}

		phrase := trig[:sep]
		count, err := strconv.Atoi(trig[sep+1:])
		if err != nil {
			return fmt.Errorf("Invalid trigger format %s around %s,\n%v", value, trig, err)
		}

		tmp[phrase] = count
	}

	*t = tmp
	return nil
}

// ConfigFromEnv method retrieves the Config object from the environment variables.
func ConfigFromEnv() (*Config, error) {
	c := &Config{}

	if err := envconfig.Process("lgtm", c); err != nil {
		return nil, err
	}

	return c, nil
}
