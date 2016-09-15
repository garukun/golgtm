package config

func NewTrigger(t map[string]int) trigger {
	return trigger(t)
}

/*
Composite struct literal mapping for testing.
*/

type ConfigGithub struct {
	Secret    string `envconfig:"secret" required:"true"`
	AuthToken string `envconfig:"auth_token" required:"true"`
	Owner     string `envconfig:"owner" required:"true"`
	Repo      string `envconfig:"repo" required:"true"`
}

type ConfigWorkflow struct {
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

type ConfigWorkflowContext struct {
	Name        string `envconfig:"name" default:"LGTM Code Review"`
	Description string `envconfig:"desc" default:"LGTM Code Review workflow."`
	URL         string `envconfig:"url" default:"https://github.com/garukun/golgtm"`
}

type ConfigWorkflowInReview struct {
	Label   string  `envconfig:"label" default:"Needs Review"`
	Trigger trigger `envconfig:"trigger" default:"ptal:1,please review:1,:-1::1"`
}

type ConfigWorkflowApproved struct {
	Label   string  `envconfig:"label" default:"Ready"`
	Trigger trigger `envconfig:"trigger" default:"lgtm:1,:+1::1"`
}
