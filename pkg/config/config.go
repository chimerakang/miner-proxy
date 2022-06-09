package config

type Config struct {
	Threads  *int      `json:"threads"`
	Name     *string   `json:"name"`
	Logger   *Logger   `json:"logger"`
	Proxy    *Proxy    `json:"proxy"`
	Api      *Api      `json:"api"`
	Debugger *Debugger `json:"debugger"`
	Secret   *Secret   `json:"secret"`
	Client   *Client   `json:"client"`
	Backend  *Backend  `json:"backend"`
}
