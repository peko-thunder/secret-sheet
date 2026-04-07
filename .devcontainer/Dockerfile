FROM golang:1.25-bookworm

# Install common dev tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /workspace
