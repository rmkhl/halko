#!/bin/bash
# Start Halko services in tmux for debugging

set -e

SESSION="halko-debug"
HALKO_DIR="/home/mte/rmkhaklab/halko"

# Default to debug level (3) if LOGLEVEL not provided
# Log levels: 0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE
LOGLEVEL=${LOGLEVEL:-3}

# Simulator configuration (if empty, uses simulator.conf default)
# Options: thermodynamic, differential, or any other simulator-*.conf file
SIMULATOR=${SIMULATOR:-}

# Check if session already exists
if tmux has-session -t "$SESSION" 2>/dev/null; then
    echo "Session '$SESSION' already exists."
    echo "Use 'tmux attach -t $SESSION' to attach to it."
    echo "Or run 'make tmux-debug-stop' to stop it first."
    exit 1
fi

# Build simulator command
SIM_CMD="./bin/simulator -loglevel $LOGLEVEL"
if [ -n "$SIMULATOR" ]; then
    SIM_CONFIG="simulator-${SIMULATOR}.conf"
    if [ ! -f "$SIM_CONFIG" ]; then
        echo "ERROR: Simulator config file '$SIM_CONFIG' not found."
        echo "Available configs:"
        ls -1 simulator*.conf 2>/dev/null || echo "  (none found)"
        exit 1
    fi
    SIM_CMD="$SIM_CMD -sim-config $SIM_CONFIG"
    echo "Creating tmux session '$SESSION' with loglevel=$LOGLEVEL, simulator=$SIMULATOR..."
else
    echo "Creating tmux session '$SESSION' with loglevel=$LOGLEVEL..."
fi

# Create new session with simulator window (detached)
tmux new-session -d -s "$SESSION" -n simulator -c "$HALKO_DIR"
tmux send-keys -t "$SESSION:simulator" "$SIM_CMD" C-m

# Create powerunit window
tmux new-window -t "$SESSION:" -n powerunit -c "$HALKO_DIR"
tmux send-keys -t "$SESSION:powerunit" "./bin/powerunit -loglevel $LOGLEVEL" C-m

# Create controlunit window
tmux new-window -t "$SESSION:" -n controlunit -c "$HALKO_DIR"
tmux send-keys -t "$SESSION:controlunit" "./bin/controlunit -loglevel $LOGLEVEL" C-m

# Create webapp window
tmux new-window -t "$SESSION:" -n webapp -c "$HALKO_DIR"
tmux send-keys -t "$SESSION:webapp" "make run-webapp" C-m

# Create a shell window for commands
tmux new-window -t "$SESSION:" -n shell -c "$HALKO_DIR"

echo "✓ Session '$SESSION' created with windows: simulator, powerunit, controlunit, webapp, shell"
echo "  Log level: $LOGLEVEL (0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE)"
if [ -n "$SIMULATOR" ]; then
    echo "  Simulator: $SIM_CONFIG"
fi
echo ""

# Check if we're inside a tmux session
if [ -n "$TMUX" ]; then
    echo "Detected running from within tmux session."
    echo "Switching to '$SESSION' session..."
    tmux switch-client -t "$SESSION"
else
    echo "To attach to the session, run:"
    echo "  tmux attach -t $SESSION"
    echo ""
    echo "To start with different settings:"
    echo "  LOGLEVEL=4 make tmux-debug-run                          # TRACE level"
    echo "  LOGLEVEL=2 make tmux-debug-run                          # INFO level"
    echo "  SIMULATOR=thermodynamic make tmux-debug-run             # Use thermodynamic model"
    echo "  LOGLEVEL=4 SIMULATOR=differential make tmux-debug-run   # Both options"
    echo ""
    echo "Tmux keybindings:"
    echo "  Ctrl+b n       - Next window"
    echo "  Ctrl+b p       - Previous window"
    echo "  Ctrl+b 0-4     - Jump to window by number (0=simulator, 1=powerunit, 2=controlunit, 3=webapp, 4=shell)"
    echo "  Ctrl+b w       - Interactive window list"
    echo "  Ctrl+b d       - Detach from session (keeps it running)"
    echo "  Ctrl+b &       - Kill current window"
    echo ""
    echo "To switch between sessions (if inside tmux):"
    echo "  Ctrl+b s       - Interactive session list"
    echo "  Ctrl+b (       - Switch to previous session"
    echo "  Ctrl+b )       - Switch to next session"
    echo ""
fi
