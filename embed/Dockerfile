FROM node:20-bookworm

RUN apt-get update && apt-get install -y \
    git ripgrep fd-find fzf jq curl ca-certificates \
    build-essential python3 python3-pip \
    less vim \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g @anthropic-ai/claude-code

RUN useradd -m -s /bin/bash cove

COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

USER cove
WORKDIR /workspace
ENTRYPOINT ["entrypoint.sh"]
CMD ["claude", "--dangerously-skip-permissions"]
