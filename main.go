package main

import (
	"fmt"
	"os"

	"github.com/JimmaaBinyamin/drone-gemini-cli-plugin/plugin"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	fmt.Println("=== Drone Gemini CLI Plugin ===")
	fmt.Println()

	var cfg plugin.Config
	if err := envconfig.Process("plugin", &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse configuration: %v\n", err)
		os.Exit(1)
	}

	p := plugin.New(cfg)

	if err := p.Exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("=== Plugin completed successfully ===")
}
