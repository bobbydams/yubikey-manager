package main

import (
	"os"

	"github.com/bobbydams/yubikey-manager/internal/cli"
)

var (
	// version is set at build time via ldflags: -X main.version=VERSION
	version = "dev"
)

func main() {
	// Set version in CLI if available
	if version != "dev" {
		cli.SetVersion(version)
	}

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
