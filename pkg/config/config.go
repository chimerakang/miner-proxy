package config

type Config struct {
	Threads *int    `json:"threads"`
	Name    *string `json:"name"`
	Logger  *Logger `json:"logger"`
	Proxy   *Proxy  `json:"proxy"`
	Api     *Api    `json:"api"`
}
