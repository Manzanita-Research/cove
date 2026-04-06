#!/bin/bash
# symlink .claude.json from the mounted .claude dir to $HOME
if [ -f "$HOME/.claude/.claude.json" ] && [ ! -f "$HOME/.claude.json" ]; then
    ln -s "$HOME/.claude/.claude.json" "$HOME/.claude.json"
fi
exec "$@"
