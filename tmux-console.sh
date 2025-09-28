#!/usr/bin/env bash
set -euo pipefail

# tmux-console.sh
# Spawns a tmux console with 3 panes in this directory:
# - Pane 0: make api
# - Pane 1: make executor
# - Pane 2: interactive shell for the user
# If the session already exists, simply attaches to it.

# Ensure tmux is installed
if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is not installed or not in PATH." >&2
  echo "Please install tmux (e.g., on macOS: brew install tmux) and re-run." >&2
  exit 1
fi

# Move to repo path of this script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Load .env if present
if [[ -f .env ]]; then
  echo "Loading .env"
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

# Session name from .env or default
SESSION="${BSS_TMUX_CONSOLE_TART_TF:-bss}"
if [[ -z "${SESSION}" ]]; then
  echo "Session name empty, aborting" >&2
  exit 1
fi

echo "Preparing tmux session '${SESSION}'..."

# If session exists, just attach
if tmux has-session -t "${SESSION}" 2>/dev/null; then
  echo "Attaching to existing tmux session '${SESSION}'..."
  exec tmux attach -t "${SESSION}"
fi

# Create new session in current directory
TMUX_START_DIR="$PWD"
tmux start-server

tmux new-session -d -s "${SESSION}" -c "${TMUX_START_DIR}" -n console

# Wait briefly for session to be recognized (up to ~3s)
for i in {1..10}; do
  if tmux has-session -t "${SESSION}" 2>/dev/null; then
    break
  fi
  sleep 0.3
done

if ! tmux has-session -t "${SESSION}" 2>/dev/null; then
  echo "Failed to create tmux session '${SESSION}'" >&2
  tmux list-sessions || true
  exit 1
fi

# Pane 0: API
tmux send-keys -t "${SESSION}:0.0" 'make api' C-m

# Pane 1: split right - Executor
tmux split-window -h -t "${SESSION}:0"
tmux send-keys -t "${SESSION}:0.1" 'make executor' C-m

# Pane 2: split bottom on right - Interactive shell for user
tmux split-window -v -t "${SESSION}:0.1"
# leave shell idle for user input

# Arrange layout and attach
tmux select-layout -t "${SESSION}:0" tiled
exec tmux attach -t "${SESSION}"
