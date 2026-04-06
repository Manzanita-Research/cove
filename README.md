# claudebox

Sandboxed Claude Code sessions using [Apple Containers](https://github.com/apple/container).

One command drops you into an ephemeral Linux microVM where Claude Code runs with `--dangerously-skip-permissions`, scoped to your current directory. Your filesystem stays safe. Claude goes feral.

## Install

```
go install github.com/manzanita-research/claudebox@latest
```

Requires [Apple's container CLI](https://github.com/apple/container/releases) installed separately. macOS 26+ on Apple silicon.

## Usage

```bash
cd ~/code/my-project
claudebox
```

That's it. First run builds the image (~30 seconds). Every run after is instant. Exit Claude to destroy the sandbox.

### What gets mounted

| Host | Container | Purpose |
|------|-----------|---------|
| Current directory | `/workspace` | Your project. The only thing Claude can see. |
| `~/.claude` | `/root/.claude` | Auth tokens, so you don't re-auth every session. |

### Flags

```
claudebox --rebuild   # force rebuild the container image
claudebox --version
claudebox --help
```

## How it works

claudebox builds a Linux container image with Node.js, Claude Code, and common dev tools (git, ripgrep, fd, jq, etc). It runs that image as a lightweight Apple Container microVM with your project mounted at `/workspace` and auto-launches `claude --dangerously-skip-permissions`. When you exit Claude, the container is destroyed.

Only your current directory is exposed. Your dotfiles, other repos, system files — all invisible inside the box.

## We built this because we needed it

`--dangerously-skip-permissions` is the best way to use Claude Code for real work, but it's dangerous on a bare filesystem. claudebox makes it safe by default.
