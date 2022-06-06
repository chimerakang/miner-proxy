package config

type Api struct {
	Enabled *bool   `json:"enabled"`
	Listen  *string `json:"listen"`
}
