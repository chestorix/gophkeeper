// internal/client/config.go
package client

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	Token         string
	DataDir       string
}

func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8090", "server address")
	flag.StringVar(&cfg.Token, "t", "", "authentication token")
	flag.StringVar(&cfg.DataDir, "d", "./data", "data directory")

	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}
	if envToken := os.Getenv("TOKEN"); envToken != "" {
		cfg.Token = envToken
	}
	if envDataDir := os.Getenv("DATA_DIR"); envDataDir != "" {
		cfg.DataDir = envDataDir
	}

	return cfg
}

func printVersion() {

	version := "dev"
	buildDate := "unknown"
	commit := "unknown"

	println("GophKeeper Client")
	println("Version:", version)
	println("Build Date:", buildDate)
	println("Commit:", commit)
}
