FROM node:20-bookworm

RUN apt-get update && apt-get install -y \
    git ripgrep fd-find fzf jq curl ca-certificates \
    build-essential python3 python3-pip \
    less vim \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g @anthropic-ai/claude-code

RUN useradd -m -s /bin/bash cove
USER cove

WORKDIR /workspace
CMD ["claude", "--dangerously-skip-permissions"]
