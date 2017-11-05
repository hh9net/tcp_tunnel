package config

type Config struct {
	ServIp   string
	ServPort string
}

func NewConfig(ip, port string) *Config {
	return &Config{
		ServIp:   ip,
		ServPort: port,
	}
}
