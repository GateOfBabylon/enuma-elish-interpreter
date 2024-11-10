package types

type Task struct {
	Name   string `yaml:"name,omitempty"`
	Export string `yaml:"export,omitempty"`

	ConditionStatements *ConditionStatements `yaml:",inline,omitempty"`
	TimeStatements      *TimeStatements      `yaml:",inline,omitempty"`
	HttpTaskFields      *HttpTaskFields      `yaml:",inline,omitempty"`
	PyTaskFields        *PyTaskFields        `yaml:",inline,omitempty"`
}

type HttpTaskFields struct {
	Method  string            `yaml:"method,omitempty"`
	Url     string            `yaml:"url,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}

type PyTaskFields struct {
	Script     string `yaml:"script,omitempty"`
	ScriptPath string `yaml:"scriptPath,omitempty"`
}
