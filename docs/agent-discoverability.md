# Making aptly Discoverable by Agents

## 1. MCP Server (highest impact)

MCP (Model Context Protocol) is the standard way agents discover and use tools.
Claude Code, Cursor, Windsurf, and Cline all support it natively.

**Goal:** Add an `aptly mcp` subcommand that speaks JSON-RPC over stdio.
Users add one line to their config and every aptly capability is auto-discovered.

```jsonc
// .mcp.json or claude_desktop_config.json
{ "mcpServers": { "aptly": { "command": "aptly", "args": ["mcp"] } } }
```

**Implementation options:**

- Rust-native: add a `mcp` subcommand using an MCP Rust crate (e.g. `rmcp`)
- Thin wrapper: a small Node/Python script that shells out to aptly (faster to ship)

**Distribution:** Publish to MCP registries like Smithery or mcp.run.

## 2. npm Distribution (npx aptly)

Publish an npm package so agents in Node environments can `npx aptly`.
Use the platform-specific optional dependencies pattern (same as esbuild, turbo, biome):

1. Publish per-platform packages: `@aptly/darwin-arm64`, `@aptly/linux-x64`, etc.
2. Each contains the pre-built Rust binary from GitHub Releases.
3. The main `aptly` package declares them as `optionalDependencies`.
4. npm auto-selects the right one at install time.

The release CI already builds for 4 platforms, so this is a matter of
adding an npm publish step.

## 3. Claude Code Skill

Re-add and maintain `.claude/skills/aptos/SKILL.md` so Claude Code
auto-discovers aptly as a domain skill. Consider having `install.sh`
optionally install the skill alongside the binary.

For Aptos projects that use aptly, a project-level `CLAUDE.md` mentioning
aptly makes Claude Code automatically aware of it in that repo context.

## 4. Package Manager Presence

| Channel        | Command                          |
| -------------- | -------------------------------- |
| Homebrew tap   | `brew install 0xbe1/tap/aptly`   |
| crates.io      | `cargo install aptly`            |
| npm            | `npx aptly`                      |

## 5. HTTP / OpenAPI Layer

For web-based agents (ChatGPT plugins, custom GPTs):

- Add `aptly serve` that exposes commands as REST endpoints
- Generate an OpenAPI spec
- Enables ChatGPT plugin registration, Composio integration, etc.

## 6. Registry Listings

- Smithery (MCP registry)
- Composio (multi-agent tool platform)
- LangChain community tools

## Recommended Order

1. Ship MCP server (covers most agent platforms at once)
2. npm distribution (easy given existing release infra)
3. Re-add Claude Code skill
4. Homebrew tap + crates.io
5. HTTP layer + registries
