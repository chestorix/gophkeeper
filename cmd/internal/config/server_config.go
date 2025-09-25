package config

type ServerConfig struct {
	RunAddress string `env:"RUN_ADDRESS"`
	DBURI      string `env:"DATABASE_URI"`
}
