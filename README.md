# nornsctl

CLI for the [Norns](https://github.com/nornscode/norns) durable agent runtime.

## Quickstart

```bash
brew install nornscode/tap/nornsctl
nornsctl dev
nornsctl new my-agent
cd my-agent
uv sync
uv run my-agent-worker
```

That's it. You have a running Norns server and a connected agent worker.

## Install

```bash
brew install nornscode/tap/nornsctl
```

Or via Go:

```bash
go install github.com/nornscode/nornsctl@latest
```

Or build from source:

```bash
git clone https://github.com/nornscode/nornsctl.git
cd nornsctl
go build -o nornsctl .
```

## Commands

### Dev server

```
nornsctl dev                                      Start local Norns server (foreground)
nornsctl dev up                                   Start in background
nornsctl dev down                                 Stop server
nornsctl dev status                               Show server status + API key
nornsctl dev logs                                 Tail server logs
nornsctl dev reset                                Stop and delete all data
```

Requires Docker. Runs Postgres and Norns in containers, generates an API key, and stores state in `~/.nornsctl/dev/`.

By default, `nornsctl dev` runs `ghcr.io/nornscode/norns:main`. To test another image:

```bash
nornsctl dev --image local/norns:dev
# or
NORNS_DEV_IMAGE=local/norns:dev nornsctl dev up
```

If pulling the default image fails with an authorization error, make sure the package is public or run `docker login ghcr.io` if you are testing a private package.

### Scaffolding

```
nornsctl new <name> [--language python] [--dir .]  Create a new agent project
```

Generates a ready-to-run agent worker project. If `nornsctl dev` is running, the project is automatically configured with the server URL and API key.

### Agents

```
nornsctl agents list                              List agents
nornsctl agents show <id>                         Show agent details
nornsctl agents create --name ... --system-prompt ... Create an agent
nornsctl agents update <id> --name ...            Update an agent
nornsctl agents status <id>                       Get agent process status
nornsctl agents message <id> --content "..."      Send a message to an agent
```

### Runs

```
nornsctl runs list [--agent <id>] [--limit N]     List runs
nornsctl runs show <id>                           Show run details + failure inspector
nornsctl runs events <id> [--json]                Print event log
nornsctl runs retry <id>                          Retry a failed run
nornsctl runs reply <id> "<answer>"               Answer a run waiting on a question
nornsctl runs tail <id>                           Stream events in real-time
```

### Triggers

```
nornsctl triggers list [--agent <id>]             List cron triggers
nornsctl triggers show <id>                       Show trigger details
nornsctl triggers create --agent <id> --name ... --cron "0 9 * * 5" --message "..."
                                                  Create a trigger (add --conversation-key
                                                  for persistent history across firings)
nornsctl triggers update <id> --cron ...          Update a trigger
nornsctl triggers enable <id>                     Enable a trigger
nornsctl triggers disable <id>                    Disable a trigger
nornsctl triggers fire <id>                       Fire now, outside the schedule
nornsctl triggers delete <id>                     Delete a trigger
```

### Gards

```
nornsctl gards list                               List gards (worker execution contexts)
nornsctl gards create [--name N] [--template T]   Create a gard; prints its claim token once
nornsctl gards inspect <id>                       Show gard details and ports
nornsctl gards ports <id>                         List a gard's registered ports
nornsctl gards destroy <id> [--force]             Destroy a gard (kicks its worker)
```

### Conversations

```
nornsctl conversations list <agent_id>            List conversations
nornsctl conversations show <agent_id> <key>      Show conversation details
nornsctl conversations delete <agent_id> <key>    Delete a conversation
```

## Configuration

```bash
export NORNS_URL=http://localhost:4000
export NORNS_API_KEY=nrn_...
```

Or via flags: `nornsctl --url http://... --api-key nrn_... agents list`

## License

MIT
