---
title: mcp2cli
description: Turn any MCP server into a CLI
template: splash
---

<div class="landing">
  <h1 class="landing-title">mcp2cli</h1>
  <p class="landing-tagline">Turn any MCP server into a command-line tool.</p>
</div>

```bash
# add a server once
mcp add time 'uvx mcp-server-time'

# use it like any command
mcp time get-current-time --timezone America/New_York
```

```bash
# discover what's available
mcp time tools
```

```text
Tools (2):

  convert-time      Convert time between timezones.
  get-current-time  Get current time in a specific timezone.
```

```bash
# inspect a tool — flags are generated from the schema
mcp time tools get-current-time
```

```text
NAME
  get-current-time - Get current time in a specific timezone

USAGE
  mcp time get-current-time --timezone <string>

ARGS
  --timezone string  Required. IANA timezone name.
```

```bash
# keep it running for instant responses
mcp time up

# share one server across terminals, Claude Desktop, notebooks
mcp time up --share
```
