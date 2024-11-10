package types

type Run struct {
	Executor *Executor `yaml:"executor"`
}

type Executor struct {
	Name     string                 `yaml:"name"`
	Type     string                 `yaml:"type"`
	Universe Universe               `yaml:"universe,omitempty"`
	Env      map[string]interface{} `yaml:"env,omitempty"`
	Tasks    []Task                 `yaml:"tasks"`
}

type Task struct {
	Name   string `yaml:"name,omitempty"`
	Export string `yaml:"export,omitempty"`

	Condition string `yaml:"condition,omitempty"`
	Iterate   string `yaml:"iterate,omitempty"`

	Delay   int `yaml:"delay,omitempty"`
	Timeout int `yaml:"timeout,omitempty"`
	Retry   int `yaml:"retry,omitempty"`

	Tries    []Try  `yaml:"try,omitempty"`
	Else     *Task  `yaml:"else,omitempty"`
	Parallel []Task `yaml:"parallel,omitempty"`

	Method  string            `yaml:"method,omitempty"`
	Url     string            `yaml:"url,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`

	Script     string `yaml:"script,omitempty"`
	ScriptPath string `yaml:"scriptPath,omitempty"`
}

type Try struct {
	If   string `yaml:"if"`
	Task *Task  `yaml:",inline"`
}

type Universe struct {
	World  string `yaml:"world"`
	Secret string `yaml:"secret"`
}
