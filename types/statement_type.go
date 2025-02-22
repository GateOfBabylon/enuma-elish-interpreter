package types

type TimeStatements struct {
	Delay   int `yaml:"delay,omitempty"`
	Timeout int `yaml:"timeout,omitempty"`
	Retry   int `yaml:"retry,omitempty"`
}

type ConditionStatements struct {
	Condition string         `yaml:"condition,omitempty"`
	Pick      *PickStatement `yaml:"pick,omitempty"`
	Parallel  *[]Task        `yaml:"parallel,omitempty"`
}

type PickStatement struct {
	IfStatement []IfStatement `yaml:"if"`
	Else        *Task         `yaml:"else,omitempty"`
}

type IfStatement struct {
	Try  string `yaml:"try,omitempty"`
	Task *Task  `yaml:",inline"`
}
