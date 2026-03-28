# mcp2cli: Turn any MCP server into a CLI

If [mcp2py](https://github.com/maximerivest/mcp2py) turns an MCP server into a Python module, `mcp2cli` turns an MCP server into a command-line tool you can use from any terminal.

## What is MCP?

MCP (Model Context Protocol) is an emerging standard that lets AI tools, APIs, and local apps describe what they can do in a machine-readable way. More and more services are exposing MCP interfaces. When they do, you don't have to learn their custom API — you just point `mcp2cli` at them and start using their tools from your terminal.

**You don't need to understand the protocol.** All you need to know is:

> If a service has an MCP server, `mcp2cli` lets you use it as if it were a normal command-line tool.

---

## Installation

`mcp2cli` is a single binary. No Go, Python, or Node.js runtime needed.

### macOS

Open **Terminal** and paste:

```bash
# Apple Silicon (M1/M2/M3/M4 — most modern Macs)
curl -L -o mcp2cli https://github.com/MaximeRivest/mcp2cli/releases/latest/download/mcp2cli-darwin-arm64
chmod +x mcp2cli
sudo mv mcp2cli /usr/local/bin/
```

<details>
<summary>Intel Mac?</summary>

```bash
curl -L -o mcp2cli https://github.com/MaximeRivest/mcp2cli/releases/latest/download/mcp2cli-darwin-amd64
chmod +x mcp2cli
sudo mv mcp2cli /usr/local/bin/
```

</details>

### Linux

Open a terminal and paste:

```bash
curl -L -o mcp2cli https://github.com/MaximeRivest/mcp2cli/releases/latest/download/mcp2cli-linux-amd64
chmod +x mcp2cli
sudo mv mcp2cli /usr/local/bin/
```

<details>
<summary>ARM64 (Raspberry Pi, etc.)?</summary>

```bash
curl -L -o mcp2cli https://github.com/MaximeRivest/mcp2cli/releases/latest/download/mcp2cli-linux-arm64
chmod +x mcp2cli
sudo mv mcp2cli /usr/local/bin/
```

</details>

### Windows

1. Download [`mcp2cli.exe`](https://github.com/MaximeRivest/mcp2cli/releases/latest/download/mcp2cli-windows-amd64.exe)
2. Rename the file to `mcp2cli.exe`
3. Put it somewhere on your `PATH`, for example `C:\Users\YourName\bin\`

Or paste this into **PowerShell**:

```powershell
# Create a bin folder if you don't have one
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin" | Out-Null

# Download
Invoke-WebRequest -Uri "https://github.com/MaximeRivest/mcp2cli/releases/latest/download/mcp2cli-windows-amd64.exe" -OutFile "$env:USERPROFILE\bin\mcp2cli.exe"

# Add to PATH for this session (to make it permanent, add it via System Settings → Environment Variables)
$env:PATH += ";$env:USERPROFILE\bin"
```

### Check that it worked

Open a **new** terminal window and run:

```bash
mcp2cli version
```

You should see a version number. If you do, you're ready.

---

## Quick start: your first MCP server

Let's try a real example. The [filesystem MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem) lets you list and read files through MCP. You'll need [Node.js](https://nodejs.org/) installed for this example (the server runs on Node, but `mcp2cli` itself doesn't need it).

### Step 1: See what the server can do

```bash
mcp2cli tools --command 'npx -y @modelcontextprotocol/server-filesystem /tmp'
```

This starts the server, connects to it, and lists all available tools. You'll see something like:

```text
list-directory    List files and directories
read-file         Read the contents of a file
write-file        Create or overwrite a file
...
```

### Step 2: Call a tool

```bash
mcp2cli tool --command 'npx -y @modelcontextprotocol/server-filesystem /tmp' list-directory /tmp
```

That calls the `list-directory` tool with `/tmp` as the path argument. You'll see the directory listing right in your terminal.

### Step 3: Register the server so you don't have to retype the command

```bash
mcp2cli add files --command 'npx -y @modelcontextprotocol/server-filesystem /tmp'
```

Now you can just write:

```bash
mcp2cli tools files
mcp2cli tool files list-directory /tmp
mcp2cli tool files read-file /tmp/hello.txt
```

### Step 4: Give it its own name

```bash
mcp2cli expose files
```

Now `mcp-files` is a real command on your system:

```bash
mcp-files tools
mcp-files list-directory /tmp
mcp-files read-file /tmp/hello.txt
```

Want an even shorter name?

```bash
mcp2cli expose files --as fs
fs list-directory /tmp
```

---

## Using a remote MCP server

Some MCP servers run on the web instead of your machine. You connect to them with a URL.

### With a bearer token (API key)

```bash
# Set your API key as an environment variable
export ACME_TOKEN="your-api-key-here"

# Use the server
mcp2cli tools --url https://api.acme.dev/mcp --bearer-env ACME_TOKEN
mcp2cli tool --url https://api.acme.dev/mcp --bearer-env ACME_TOKEN search --query invoices
```

Or register it once:

```bash
mcp2cli add acme --url https://api.acme.dev/mcp --bearer-env ACME_TOKEN
mcp2cli tools acme
mcp2cli tool acme search --query invoices
```

### With OAuth (browser login)

```bash
mcp2cli add notion --url https://mcp.notion.com/mcp --auth oauth
mcp2cli login notion
# Your browser opens, you log in, come back to the terminal

mcp2cli tool notion notion-get-self
```

---

## Inspecting tools before calling them

You can look at any tool's arguments before calling it:

```bash
mcp2cli tools weather get-forecast
```

Output:

```text
NAME
  get-forecast - Get weather forecast for a location

USAGE
  mcp2cli tool weather get-forecast --latitude <float> --longitude <float>

ARGS
  --latitude float   Required. Latitude of the location.
  --longitude float  Required. Longitude of the location.
```

---

## Resources and prompts

MCP servers can also expose **resources** (data you can read) and **prompts** (text templates you can render).

```bash
# List resources
mcp2cli resources weather

# Read one
mcp2cli resource weather api-docs

# List prompts
mcp2cli prompts weather

# Render a prompt with arguments
mcp2cli prompt weather review-code --code 'x <- 1' --focus api
```

---

## Interactive shell mode

If you want to explore a server interactively without retyping the server name every time:

```bash
mcp2cli shell weather
```

```text
weather> tools
weather> get-forecast --latitude 37.7 --longitude -122.4
weather> resources
weather> resource api-docs
weather> set output json
weather> get-forecast --latitude 37.7 --longitude -122.4
weather> exit
```

The shell keeps the connection open, supports history, tab completion, and lets you switch output formats on the fly.

---

## Output formats

By default, `mcp2cli` prints human-friendly output. When you need machine-readable data for scripting:

```bash
# Human-readable (default)
mcp2cli tool weather get-forecast 37.7 -122.4

# Exact JSON for scripts
mcp2cli tool weather get-forecast 37.7 -122.4 -o json

# YAML
mcp2cli tool weather get-forecast 37.7 -122.4 -o yaml

# Plain text only
mcp2cli resource weather api-docs -o raw
```

The `-o json` flag is always script-safe: output goes to `stdout`, diagnostics go to `stderr`, and exit codes are stable.

---

## Arguments: flags and positionals

`mcp2cli` reads the tool's schema and generates CLI flags automatically.

```bash
# Named flags (always work)
mcp2cli tool weather get-forecast --latitude 37.7 --longitude -122.4

# Positional arguments (for required scalar args, in schema order)
mcp2cli tool weather get-forecast 37.7 -122.4

# Booleans
mcp2cli tool api update --dry-run
mcp2cli tool api update --no-dry-run

# Repeated values for arrays
mcp2cli tool api search --tag cli --tag go --tag mcp

# Structured JSON from a file
mcp2cli tool api create --payload @data.json

# Or from stdin
cat data.json | mcp2cli tool api create --payload @-
```

For tools with very complex schemas, you can always fall back to raw JSON:

```bash
mcp2cli tool api complex-tool --input '{"nested": {"key": "value"}}'
mcp2cli tool api complex-tool --input @payload.json
```

---

## Diagnosing problems

If something isn't working:

```bash
mcp2cli doctor weather
```

```text
CHECK    STATUS  DETAIL
resolve  ok      weather
command  ok      /usr/local/bin/npx
auth     ok      no auth required
connect  ok      initialize handshake succeeded
tools    ok      2 tool(s) available
```

---

## Config

Servers you register are saved automatically.

- Global config: `~/.config/mcp2cli/config.yaml`
- Per-project config: `.mcp2cli.yaml` in the current directory

```yaml
version: 1

servers:
  weather:
    command: npx -y @h1deya/mcp-server-weather
    expose:
      - mcp-weather

  notion:
    url: https://mcp.notion.com/mcp
    auth: oauth
```

You manage servers with:

```bash
mcp2cli add weather --command '...'
mcp2cli ls
mcp2cli rm weather
```

---

## Shell completions

Enable tab completion for your shell:

```bash
# bash
echo 'source <(mcp2cli completion bash)' >> ~/.bashrc

# zsh
echo 'source <(mcp2cli completion zsh)' >> ~/.zshrc

# fish
mcp2cli completion fish | source
```

Exposed commands like `mcp-weather` and `wea` also support completions:

```bash
echo 'source <(mcp-weather completion bash)' >> ~/.bashrc
```

---

## Complete command reference

| Command | What it does |
| --- | --- |
| `mcp2cli add <name>` | Register a server |
| `mcp2cli ls` | List registered servers |
| `mcp2cli rm <name>` | Remove a registered server |
| `mcp2cli expose <name>` | Create a standalone command for a server |
| `mcp2cli unexpose <name>` | Remove an exposed command |
| `mcp2cli login <name>` | Authenticate ahead of time |
| `mcp2cli tools <server> [tool]` | List tools or inspect one |
| `mcp2cli tool <server> <tool> [args...]` | Call a tool |
| `mcp2cli resources <server> [resource]` | List or inspect resources |
| `mcp2cli resource <server> <resource>` | Read a resource |
| `mcp2cli prompts <server> [prompt]` | List or inspect prompts |
| `mcp2cli prompt <server> <prompt> [args...]` | Render a prompt |
| `mcp2cli shell <server>` | Open interactive shell |
| `mcp2cli doctor <server>` | Diagnose connection issues |
| `mcp2cli completion <shell>` | Generate shell completions |
| `mcp2cli version` | Print version |

---

## How `mcp2cli` relates to `mcp2py`

[mcp2py](https://github.com/maximerivest/mcp2py) turns MCP servers into **Python modules** — great for notebooks, scripts, and data analysis in Python.

`mcp2cli` turns MCP servers into **shell commands** — great for terminal users, shell scripts, CI pipelines, and anyone who prefers the command line.

They complement each other. If a service has an MCP server:
- use `mcp2py` to call it from Python
- use `mcp2cli` to call it from bash, zsh, fish, PowerShell, or any terminal

---

## Build from source

If you have [Go](https://go.dev/dl/) installed:

```bash
git clone https://github.com/MaximeRivest/mcp2cli.git
cd mcp2cli
go build -o mcp2cli ./cmd/mcp2cli
./mcp2cli version
```

---

## Current status

This is an alpha release. What works today:

- ✅ local stdio MCP servers
- ✅ remote HTTP JSON-RPC MCP servers
- ✅ bearer token and custom header auth
- ✅ OAuth login with browser flow and token persistence
- ✅ tools, resources, and prompts
- ✅ schema-driven CLI flags and positional arguments
- ✅ exposed server commands (`mcp-weather`, `wea`, etc.)
- ✅ interactive shell mode with history and completion
- ✅ terminal elicitation (server-initiated user prompts)
- ✅ metadata cache for fast completions
- ✅ `doctor` diagnostics

Still coming:

- ⬜ sampling (LLM-backed server requests)
- ⬜ SSE / streamable HTTP transport compatibility

---

If `mcp2py` makes MCP feel like a native Python library, `mcp2cli` makes MCP feel like it was built for the terminal from day one.
