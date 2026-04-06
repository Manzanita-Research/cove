# cove

Sandboxed Claude Code sessions using [Apple Containers](https://github.com/apple/container).

One command drops you into an ephemeral Linux microVM where Claude Code runs with `--dangerously-skip-permissions`, scoped to your current directory. Your filesystem stays safe. Claude goes feral.

## Requirements

- macOS 26+ (Tahoe) on Apple silicon
- [Apple's container CLI](https://github.com/apple/container/releases)
- Go 1.21+ (for installation)
- An authenticated Claude Code session on the host (Pro, Max, Team, Enterprise, or API)

## Install

```
go install github.com/manzanita-research/cove@latest
```

## Usage

```bash
cd ~/code/my-project
cove
```

First run takes ~30 seconds to build the container image. Every run after that is instant. Exit Claude to destroy the sandbox.

### What happens

1. Checks for the `container` CLI and starts the container system if needed
2. Builds the `cove:latest` image on first run (Node.js 20, Claude Code, dev tools)
3. Extracts your Claude OAuth credentials from macOS Keychain
4. Mounts your project and auth into the container
5. Launches `claude --dangerously-skip-permissions`
6. On exit, the container is destroyed

### What gets mounted

| Host | Container | Why |
|------|-----------|-----|
| Current directory | `/workspace` | Your project. The only thing Claude can see. |
| `~/.claude` | `/home/cove/.claude` | Auth and config, so you don't re-auth every session. |

Everything else on your machine — dotfiles, other repos, system files — is invisible inside the cove.

### Flags

```
cove              # launch a sandboxed session for the current directory
cove --rebuild    # force rebuild the container image
cove --version
cove --help
```

### Tools available inside the container

git, ripgrep, fd, fzf, jq, curl, vim, python3, build-essential, and whatever's in Node.js 20.

## How it works

cove is a thin Go wrapper around Apple's `container` CLI. The Dockerfile is embedded in the binary via `go:embed` — no files to lose, no config to manage.

Authentication is the tricky part. On macOS, Claude Code stores OAuth tokens in the Keychain. The Linux container expects them at `~/.claude/.credentials.json`. cove extracts the credentials from your Keychain on each launch and writes them where the container can find them.

The container runs as a non-root user (`cove`) because Claude Code refuses `--dangerously-skip-permissions` as root.

## We built this because we needed it

`--dangerously-skip-permissions` is the best way to use Claude Code for real work, but it's dangerous on a bare filesystem. cove makes it safe by default.
