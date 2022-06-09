package config

type Secret struct {
	Enable *bool   `json:"enable"`
	Key    *string `json:"key"`
}
