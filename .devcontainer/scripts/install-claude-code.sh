#!/bin/bash
set -e

# Setup symlink for .claude.json to persist settings in volume-mounted directory
# This ensures that Claude Code settings (theme, tips history, etc.) are preserved across container rebuilds
setup_claude_config_persistence() {
  local claude_dir="/home/vscode/.claude"
  local settings_in_volume="$claude_dir/config.json"
  local settings_link="/home/vscode/.claude.json"


  # If .claude.json exists as a regular file, move it to volume
  if [ -f "$settings_link" ] && [ ! -L "$settings_link" ]; then
    mv "$settings_link" "$settings_in_volume"
  fi

  # Create initial config.json in volume if it doesn't exist
  if [ ! -f "$settings_in_volume" ]; then
    echo '{}' > "$settings_in_volume"
  fi

  # Create symlink from ~/.claude.json to ~/.claude/config.json
  if [ -L "$settings_link" ] || [ -f "$settings_link" ]; then
    rm -f "$settings_link"
  fi
  ln -s "$settings_in_volume" "$settings_link"
}

# Avoid permission error in installing
# This directory is already created with mounts property in devcontainer.json, but sometimes its owner is root
sudo chown -R $USER:$USER /home/vscode/.claude

setup_claude_config_persistence

# Install Claude Code
# refs: https://docs.claude.com/en/docs/claude-code/setup
curl -fsSL https://claude.ai/install.sh | bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
