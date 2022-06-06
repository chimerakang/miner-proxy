package config

type Logger struct {
	Level *int    `json:"level"`
	Mode  *string `json:"mode"`
	File  *string `json:"file"`
}
