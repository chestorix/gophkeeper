package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	RunAddress string `env:"RUN_ADDRESS"`
	DBURI      string `env:"DATABASE_URI"`
	JWTSecret  string `env:"JWT_SECRET"`
}

func Load() *ServerConfig {
	cfg := &ServerConfig{}
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8090", "address and port to run server")
	flag.StringVar(&cfg.DBURI, "d", "", "host=<host> user=<user> password=<password> dbname=<dbname> sslmode=<disable/enable>")
	flag.StringVar(&cfg.JWTSecret, "j", "secret", "JWT secret key")
	flag.Parse()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	}
	if envDBURI := os.Getenv("DATABASE_URI"); envDBURI != "" {
		cfg.DBURI = envDBURI
	}
	if envJWTSecret := os.Getenv("JWT_SECRET"); envJWTSecret != "" {
		cfg.JWTSecret = envJWTSecret
	}
	return cfg
}
