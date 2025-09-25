package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	RunAddress string `env:"RUN_ADDRESS"`
	DBURI      string `env:"DATABASE_URI"`
}

func Load() *ServerConfig {
	cfg := &ServerConfig{}
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8090", "address and port to run server")
	flag.StringVar(&cfg.DBURI, "d", "", "host=<host> user=<user> password=<password> dbname=<dbname> sslmode=<disable/enable>")
	flag.Parse()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	}
	if envDBURI := os.Getenv("DATABASE_URI"); envDBURI != "" {
		cfg.DBURI = envDBURI
	}
	return cfg
}
