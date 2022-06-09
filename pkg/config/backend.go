package config

type Backend struct {
	Enable  *bool   `json:"enable"`
	Listen  *string `json:"listen"`
	Pool    *string `json:"pool"`
	Web     *bool   `json:"web"`
	Offline *int    `json:"offline"`
}
