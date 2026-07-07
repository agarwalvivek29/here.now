// Command herenow is the single binary (CLI + server) for the here.now artifact
// host. Subcommands: login, publish, ls, serve. See services/herenow-api/AGENTS.md.
package main

import (
	"fmt"
	"os"

	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
