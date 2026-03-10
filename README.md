<div style="text-align: center;"><img src="./assets/icon.svg" width="80" alt="hi"></div>

[![Build Status](https://github.com/jimschubert/hi/actions/workflows/build.yml/badge.svg)](https://github.com/jimschubert/hi/actions/workflows/build.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jimschubert/hi)](https://github.com/jimschubert/hi/blob/main/go.mod)
[![License](https://img.shields.io/github/license/jimschubert/hi?a=b&color=blue)](https://github.com/jimschubert/hi/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimschubert/hi)](https://goreportcard.com/report/github.com/jimschubert/hi)
[![GitHub Release](https://img.shields.io/github/v/release/jimschubert/hi)](https://github.com/jimschubert/hi/releases/latest)

> A single notification system for every agent.

`hi` ("human intelligence") is a work-in-progress persistent desktop daemon which provides AI agents (Claude, Qwen, Copilot, and non-alliterative others) a single point of human interaction.
As an MCP server, this avoids ending a session where possible which _might_ save you credits/requests.

Here's how it works.

1. Any agent can send a `hi_*` request to the `hi` daemon
2. The daemon will display a notification to the user that the agent is awaiting human input
3. The human (you) clicks on the "Show next request" in your system tray
4. `hi` presents you with a UI for text input, single or multiple choice, confirm, etc. and you respond to the request

## Disclaimer

`hi` is a tool for human-in-the-loop agent interactions. Do **not** rely on agents for critical decisions without human review.

I'm making this tool available as-is, just please don't use it to generate "slop".

See [Anthropic's Framework for Developing Safe and Trustworthy Agents](https://www.anthropic.com/news/our-framework-for-developing-safe-and-trustworthy-agents)
for details on responsible agentic AI.

>[!WARNING]
> Be aware that some agents may trigger automatic bans if you're running long-running tasks of many steps.
> For instance, see [this discussion topic](https://github.com/orgs/community/discussions/165798).

## Install

```shell
go install github.com/jimschubert/hi@latest
```

## Usage

The `hi` binary runs the daemon, processes MCP request over stdio, and manages the daemon over IPC.

```shell
$ hi --help                       
Usage: hi <command> [flags]

Human Intelligence — MCP server for human-in-the-loop agent interactions

Flags:
  -h, --help       Show context-sensitive help.
  -v, --version    Print version information.

Commands:
  daemon      Run the hi daemon
  status      Query status of the hi daemon
  ping        Ping the hi daemon
  stdio       Query the daemon via stdio over Unix socket
  shutdown    Shutdown the hi daemon (gracefully)
  test        Submit 1+ test requests to the hi daemon over Unix socket

Run "hi <command> --help" for more information on a command.
```

## MCP configuration

You can serve MCP over IPC (Unix socket), HTTP, or both. The default is IPC because it's the easies to get started.

### IPC (Unix Socket)

This is all you need… the `stdio` subcommand will automatically start the daemon in IPC mode.

```json
{
  "mcpServers": {
    "human-intelligence": {
      "command": "/path/to/hi",
      "args": [
        "stdio"
      ]
    }
  }
}
```

### HTTP

Assuming you've already started `hi daemon` with `--addr localhost:45678`:

```json
{
  "mcpServers": {
    "human-intelligence": {
      "url": "http://localhost:45678/mcp"
    }
  }
}
```

## Example Prompt

Here's an example system prompt to ask your LLM to use `hi`:

```
When you need user input, decisions, or confirmations, use the human-intelligence MCP server tools
(get_user_input, get_user_choice, show_confirmation_dialog, etc.)
instead of asking me in chat.

This keeps our conversation clean and provides better interaction.
```

## License

Apache 2.0 - see [LICENSE](LICENSE)
