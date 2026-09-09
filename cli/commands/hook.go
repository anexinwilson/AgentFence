package commands

import (
	"fmt"
	"os"

	"github.com/agentfence/agentfence/adapters/protocol"
	"github.com/agentfence/agentfence/core"
)

// RunHook handles lifecycle hook events invoked by Antigravity on stdin/stdout.
func RunHook(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: hook subcommand required ('pre-tool' or 'stop')")
		os.Exit(1)
	}

	hookType := args[0]
	cfg, err := core.Load("")
	if err != nil {
		cfg = core.DefaultConfig("")
	}

	cwd, _ := os.Getwd()

	switch hookType {
	case "pre-tool":
		if err := protocol.HandlePreTool(os.Stdin, os.Stdout, cfg.Guardrails); err != nil {
			fmt.Fprintf(os.Stderr, "Error handling pre-tool hook: %v\n", err)
			os.Exit(1)
		}
	case "stop":
		if err := protocol.HandleStop(os.Stdin, os.Stdout, cfg.Guardrails, cwd); err != nil {
			fmt.Fprintf(os.Stderr, "Error handling stop hook: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown hook type '%s'. Expected 'pre-tool' or 'stop'.\n", hookType)
		os.Exit(1)
	}
}
