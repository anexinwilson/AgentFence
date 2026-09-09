package cli

import (
	"fmt"
	"os"

	"github.com/agentfence/agentfence/cli/commands"
	"github.com/agentfence/agentfence/cli/ui"
)

// Execute runs the main CLI application.
func Execute(args []string) {
	if len(args) < 2 {
		printUsage()
		return
	}

	command := args[1]
	subArgs := args[2:]

	switch command {
	case "init":
		commands.RunInit(subArgs)
	case "check":
		commands.RunCheck(subArgs)
	case "verify":
		commands.RunVerify(subArgs)
	case "evidence":
		commands.RunEvidence(subArgs)
	case "doctor":
		commands.RunDoctor(subArgs)
	case "explain":
		commands.RunExplain(subArgs)
	case "hook":
		commands.RunHook(subArgs)
	case "stop", "pause", "disable":
		commands.RunStop(subArgs)
	case "start", "resume", "enable":
		commands.RunStart(subArgs)
	case "version", "--version", "-v":
		fmt.Println("AgentFence v0.1.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		ui.Fail("Unknown command '%s'", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(ui.Banner())
	fmt.Print(`Usage:
  agentfence <command> [flags]

Commands:
  init       Initialize AgentFence guardrails and Antigravity hooks
  check      Perform a fast health and guardrails check
  verify     Run deterministic verification pipeline (tests, typecheck, requirements)
  evidence   Generate and display the clean, objective task evidence report
  doctor     Inspect repository, toolchain, and Antigravity hook integration health
  stop       Temporarily pause AgentFence protection hooks
  start      Resume AgentFence protection and re-enable hooks
  explain    Explain an active guardrail rule and how to modify or resolve it
  hook       Internal hook handler invoked by Antigravity (pre-tool, stop)
  version    Show version

Flags:
  --help, -h        Show help for command
  --markdown, -m    Output evidence in clean markdown format
  --json, -j        Output evidence in JSON format
`)
}
