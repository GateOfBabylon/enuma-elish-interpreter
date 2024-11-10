package types

type Run struct {
	Executor *Executor `yaml:"executor"`
}

type Executor struct {
	Name     string                 `yaml:"name"`
	Type     string                 `yaml:"type"`
	Universe *Universe              `yaml:"universe,omitempty"`
	Env      map[string]interface{} `yaml:"env,omitempty"`
	Tasks    []Task                 `yaml:"tasks"`
}

type Universe struct {
	World  string `yaml:"world"`
	Secret string `yaml:"secret"`
}
