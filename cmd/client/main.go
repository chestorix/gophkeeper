// cmd/client/main.go
package main

import (
	"fmt"
	"os"

	"github.com/chestorix/gophkeeper/internal/client"
	"github.com/chestorix/gophkeeper/internal/ui"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	cfg := client.LoadConfig()

	if cfg.Token == "" {
		cfg.Token = ui.LoadToken(cfg.DataDir)
	}

	client := client.NewClient(cfg)
	cli := ui.NewCLI(client, cfg)

	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
