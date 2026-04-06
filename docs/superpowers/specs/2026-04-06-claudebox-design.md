# claudebox — design spec

A single command that drops you into a sandboxed Claude Code session for whatever directory you're in. Uses Apple Containers (native macOS microVMs) to isolate Claude with `--dangerously-skip-permissions` so it can go feral without touching your real filesystem.

## Installation

```
go install github.com/manzanita-research/claudebox@latest
```

Requires the Apple `container` CLI ([apple/container](https://github.com/apple/container)) installed separately. macOS 26+ on Apple silicon.

## CLI

Built with Cobra for discoverability (other agents can read `--help`).

```
claudebox              # run a sandboxed claude session for cwd
claudebox --rebuild    # force rebuild the image
claudebox --version    # print version
claudebox --help       # auto-generated
```

No subcommands in v1. Just flags on the root command.

## Runtime flow

When you run `claudebox`:

1. **Check for `container` CLI** — if not on PATH, print install instructions pointing to apple/container GitHub releases and exit.
2. **Check container system status** — run `container system status`. If not running, start it with `container system start`.
3. **Check if `claudebox:latest` image exists** — run `container image list`, grep for `claudebox`. If missing (or `--rebuild`), build it.
4. **Build image** — write the embedded Dockerfile to a temp dir, run `container build -t claudebox:latest <tmpdir>`, clean up the temp dir.
5. **Print entry banner** — project name, what's mounted, how to exit. Warm colors.
6. **Run the container:**
   ```
   container run -it --rm \
     --name claudebox-<project>-<pid> \
     -v <cwd>:/workspace \
     -v ~/.claude:/root/.claude \
     -w /workspace \
     claudebox:latest \
     claude --dangerously-skip-permissions
   ```
7. **Exit claude = exit container = sandbox destroyed.**

## Embedded Dockerfile

Baked into the Go binary via `go:embed`. Single artifact, nothing to lose.

```dockerfile
FROM node:20-bookworm

RUN apt-get update && apt-get install -y \
    git ripgrep fd-find fzf jq curl ca-certificates \
    build-essential python3 python3-pip \
    less vim \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g @anthropic-ai/claude-code

WORKDIR /workspace
CMD ["claude", "--dangerously-skip-permissions"]
```

Tools included: git, ripgrep, fd, fzf, jq, curl, build-essential, python3, vim. These are what Claude Code commonly needs when working in a project.

## Project structure

```
claudebox/
  main.go                # entry point
  cmd/
    root.go              # cobra root command, main logic
  internal/
    container/
      container.go       # wraps calls to `container` CLI
    banner/
      banner.go          # colored terminal output
  embed/
    Dockerfile           # embedded into binary
  go.mod
  go.sum
```

### `internal/container/container.go`

Functions that shell out to the `container` CLI:

- `SystemStatus() error` — check if container system is running
- `SystemStart() error` — start the container system
- `ImageExists(name string) (bool, error)` — check if an image exists
- `Build(tag string, contextDir string) error` — build an image from a Dockerfile
- `Run(opts RunOpts) error` — run an interactive container with volumes, name, working dir

All functions use `os/exec.Command`, connect stdin/stdout/stderr to the terminal for interactive use, and return clear errors.

### `internal/banner/banner.go`

Simple colored output using ANSI escape codes. Two functions:

- `Warm(msg string)` — warm amber text (ANSI 215)
- `Dim(msg string)` — muted gray text (ANSI 245)

No dependencies. No lipgloss. Just printf with escape codes.

### `cmd/root.go`

The Cobra root command. Orchestrates the flow:

1. Parse flags (`--rebuild`)
2. Check for `container` binary
3. Ensure system is running
4. Build image if needed
5. Print banner
6. Exec into the container

### `main.go`

Just calls `cmd.Execute()`.

## Mounts

| Host path | Container path | Purpose |
|-----------|---------------|---------|
| `$(pwd)` | `/workspace` | Your project — the only directory Claude can see |
| `~/.claude` | `/root/.claude` | Auth tokens — shared so you don't re-auth every session |

Only the current directory is exposed. Your dotfiles, other repos, everything else — invisible to Claude inside the box.

## Error handling

- `container` not found → print: "claudebox requires Apple's container CLI. Install it from https://github.com/apple/container/releases"
- `container system start` fails → print the error output and suggest checking Apple Containers setup
- Image build fails → stream build output to terminal so user can see what went wrong, suggest `--rebuild`
- All `os/exec` errors surfaced with context, never swallowed

## What's NOT in v1

Future features, left out to keep v1 clean:

- `--persist <name>` — named containers that survive exit
- `--fresh-auth` — skip mounting `~/.claude`
- `--skill <path>` — mount a skill directory read-only
- `--no-network` — restrict egress to `api.anthropic.com` only
- `--dotfiles <repo>` — run `chezmoi init --apply` inside the container
- `--shell` — drop into bash instead of auto-launching claude
- Multi-runtime support (no Docker, no Podman)
- TUI beyond colored text (no bubbletea)

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- That's it. Everything else is stdlib.
