# nornsctl

CLI for the [Norns](https://github.com/amackera/norns) durable agent runtime.

## Quickstart

```bash
brew install amackera/tap/nornsctl
nornsctl dev
nornsctl new my-agent
cd my-agent
uv sync
uv run my-agent-worker
```

That's it. You have a running Norns server and a connected agent worker.

## Install

```bash
brew install amackera/tap/nornsctl
```

Or via Go:

```bash
go install github.com/amackera/nornsctl@latest
```

Or build from source:

```bash
git clone https://github.com/amackera/nornsctl.git
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
nornsctl runs tail <id>                           Stream events in real-time
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
