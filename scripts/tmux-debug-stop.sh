#!/bin/bash
# Stop Halko debug tmux session

set -e

SESSION="halko-debug"

# Check if session exists
if ! tmux has-session -t "$SESSION" 2>/dev/null; then
    echo "Session '$SESSION' does not exist."
    exit 0
fi

# Check if we're running from within the session we're trying to kill
if [ -n "$TMUX" ]; then
    CURRENT_SESSION=$(tmux display-message -p '#S')
    if [ "$CURRENT_SESSION" = "$SESSION" ]; then
        echo "ERROR: Cannot stop session '$SESSION' from within itself."
        echo ""
        echo "You are currently in the '$SESSION' session."
        echo "Please switch to another session first, then run: make tmux-debug-stop"
        echo ""
        echo "To switch sessions:"
        echo "  Ctrl+b s       - Interactive session list (arrow keys + Enter)"
        echo "  Ctrl+b (       - Switch to previous session"
        echo "  Ctrl+b )       - Switch to next session"
        echo ""
        echo "Or create a new session:"
        echo "  Ctrl+b :       - Command prompt, then type: new-session"
        exit 1
    fi
fi

echo "Stopping session '$SESSION'..."

# Kill all windows in the session (this will send SIGTERM to all processes)
# The -a flag kills all windows except the current one, but we'll kill the session anyway
echo "Sending termination signals to all processes..."
tmux list-windows -t "$SESSION" -F "#{window_index}" | while read -r window; do
    tmux send-keys -t "$SESSION:$window" C-c 2>/dev/null || true
done

# Give processes a moment to terminate gracefully
sleep 1

# Kill the entire session
echo "Terminating session..."
tmux kill-session -t "$SESSION"

echo "✓ Session '$SESSION' stopped successfully"
