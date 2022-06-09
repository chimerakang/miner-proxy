package config

type Debugger struct {
	Enable *bool   `json:"enable"`
	Listen *string `json:"listen"`
}
