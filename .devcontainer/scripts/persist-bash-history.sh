#!/bin/bash
set -e

sudo mkdir -p /commandhistory
sudo chown -R $USER:$USER /commandhistory
touch /commandhistory/.bash_history
echo "export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" >> ~/.bashrc

# this script works with mounts property in devcontainer.json
# refs: https://code.visualstudio.com/remote/advancedcontainers/persist-bash-history
