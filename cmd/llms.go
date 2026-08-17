package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var llmsCmd = &cobra.Command{
	Use:   "llms",
	Short: "Print usage documentation written for LLM agents",
	Long: `Prints a complete, self-contained guide to nornsctl in Markdown,
written for LLM coding agents. Pipe it into an agent's context to teach
it the CLI in one shot: nornsctl llms | pbcopy`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(llmsDoc)
	},
}

func init() {
	rootCmd.AddCommand(llmsCmd)
}

const llmsDoc = `# nornsctl — CLI for the Norns durable agent runtime

Norns is a durable agent runtime: agents are persistent state machines
that survive crashes and replay from an event log. The orchestrator never
executes LLM calls or tools itself — connected *workers* do, holding all
credentials. nornsctl is the operator CLI: inspect agents and runs, read
event logs, manage triggers/hooks/gards, scaffold worker projects, and
run a local dev server.

## Configuration

Every command needs the API location and key, from flags or environment:

- ` + "`--url` / `NORNS_URL`" + ` — API base URL (default http://localhost:4000)
- ` + "`--api-key` / `NORNS_API_KEY`" + ` — bearer token (starts with nrn_)

If a command fails with connection refused, the server at NORNS_URL is
not running — try ` + "`nornsctl dev status`" + ` or ask the operator.

## Core concepts

- **Agent** — a durable, addressable state machine with a system prompt
  and model. States: idle, awaiting_llm, awaiting_tools, waiting (for a
  human answer). Referenced by numeric ID.
- **Run** — one unit of agent work, with a versioned event log. Runs can
  fail, be retried, or block on a question (ask_human).
- **Conversation** — persistent chat history for an agent, keyed by an
  external string; reuse the key for multi-turn.
- **Worker** — an external process (usually built with a Norns SDK) that
  connects over WebSocket and executes LLM calls and tools.
- **Gard** — a worker execution context with filesystem affinity, for
  coding-agent workloads. Most workers need no gard.
- **Trigger** — a cron schedule that starts a run.
- **Hook** — an inbound webhook that starts a run from an HTTP payload.

## Commands

### Send a message and read the reply (most common operation)

    nornsctl agents message <id> --content "..." --wait

The message body goes in --content (NOT a positional argument). --wait
blocks until the run completes, fails, or asks a question, then prints
the result; --timeout <secs> adjusts the wait (default 120). Add
--conversation-key <key> for multi-turn context, and --gard <id> to
bind the run to a gard (ALL of its dispatch — LLM and tools — then goes
only to that gard's worker). Prints "Run ID: N" — use that ID with the
runs commands.

If the run asks a question, answer it with:

    nornsctl runs reply <run-id> "<answer>"

### Debug a failing run

    nornsctl runs list [--agent <id>] [--limit N]   # find the run
    nornsctl runs show <id>                          # status + failure inspector
    nornsctl runs events <id> --json                 # full event log
    nornsctl runs tail <id>                          # live event stream
    nornsctl runs retry <id>                         # retry a failed run

Start with ` + "`runs show`" + ` (human-readable failure summary), then
` + "`runs events --json`" + ` for the full picture. A run stuck in awaiting_llm
or awaiting_tools with no progress usually means no worker is connected
— check ` + "`agents status <id>`" + `.

### Agents

    nornsctl agents list
    nornsctl agents show <id>
    nornsctl agents status <id>      # live process state
    nornsctl agents create --name <n> --system-prompt <p> [--model <m>] [--purpose <p>]
    nornsctl agents update <id> [same flags]

### Triggers (cron)

    nornsctl triggers list [--agent <id>]
    nornsctl triggers create --agent <id> --name <n> --cron "0 9 * * 1" --message "..." [--conversation-key <k>] [--disabled]
    nornsctl triggers fire <id>      # fire now, outside the schedule
    nornsctl triggers enable|disable|delete <id>

### Hooks (inbound webhooks)

    nornsctl hooks list
    nornsctl hooks create --agent <id> --name <n> [--signature github|stripe|slack|none] [--signing-secret <s>] [--message-path <dot.path>] [--conversation-key-path <dot.path>]
    nornsctl hooks enable|disable|delete <id>

create prints the delivery URL once — capture it from the output.

### Gards (worker execution contexts)

    nornsctl gards list
    nornsctl gards create [--name <n>]    # prints the claim token ONCE — capture it
    nornsctl gards inspect <id>
    nornsctl gards ports <id>
    nornsctl gards destroy <id>           # soft-delete; kicks its worker

### Conversations

    nornsctl conversations list <agent-id>
    nornsctl conversations show <agent-id> <key>
    nornsctl conversations delete <agent-id> <key>

### Scaffold a worker project

    nornsctl new <name> [--template default|slack-bot] [--language python] [--dir <path>]

Generates a uv-based Python project (pyproject.toml, Dockerfile,
worker.py, tools.py) using the norns-sdk. After scaffolding: cd in, put
NORNS_API_KEY and any tool secrets in .env, and run with
` + "`uv run python -m <name>.worker`" + `.

### Local dev server

    nornsctl dev up        # start in background (Docker)
    nornsctl dev status
    nornsctl dev logs
    nornsctl dev down
    nornsctl dev reset     # stop AND delete all data — destructive

## Conventions and gotchas

- IDs are numeric and shown in every list command.
- ` + "`--json`" + ` exists where noted (runs events); other output is tabular.
- Exit code is non-zero on API errors; error text goes to stderr.
- Secrets (API keys, signing secrets, claim tokens) are printed once at
  creation and never again — never echo them into logs or commits.
`
