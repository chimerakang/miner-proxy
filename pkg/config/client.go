package config

type Client struct {
	Enable     *bool   `json:"enable"`
	Listen     *string `json:"listen"`
	MaxConnect *int    `json:"maxConnect"`
	Pools      *string `json:"pools"`
	Backend    *string `json:"backend"`
}
