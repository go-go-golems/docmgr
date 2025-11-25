#!/usr/bin/env bash
set -euo pipefail

# Reusable tmux helper: sets up a session with bash (left) and zsh (right),
# sources dynamic carapace snippets, sends actual TAB keys to trigger completion,
# and captures the pane output to /tmp files for inspection.
#
# Usage:
#   ./01-tmux-completion-test.sh [SESSION_NAME]
#   SESSION=dctest ./01-tmux-completion-test.sh
#
# Outputs (default):
#   /tmp/dctest_bash_doc_type.txt
#   /tmp/dctest_bash_status.txt
#   /tmp/dctest_bash_intent.txt
#   /tmp/dctest_bash_topics.txt
#   /tmp/dctest_bash_ticket.txt
#   /tmp/dctest_zsh_doc_type.txt
#   /tmp/dctest_zsh_status.txt
#   /tmp/dctest_zsh_intent.txt
#   /tmp/dctest_zsh_topics.txt
#   /tmp/dctest_zsh_ticket.txt

SESSION="${1:-${SESSION:-dctest}}"
ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
DOCMGR_DIR="${ROOT}/docmgr"
DIST="${DOCMGR_DIR}/dist"

# Build latest docmgr
(
  cd "${DOCMGR_DIR}"
  go build -o "${DIST}/docmgr" ./cmd/docmgr
)

# Reset session
if tmux has-session -t "${SESSION}" 2>/dev/null; then
  tmux kill-session -t "${SESSION}"
fi

# Create session with bash pane, then split to zsh
tmux new-session -d -s "${SESSION}" -x 160 -y 48 "bash"
tmux split-window -h -t "${SESSION}:0.0" "zsh"

# Setup PATH and source carapace snippets
tmux send-keys -t "${SESSION}:0.0" "export PATH=${DIST}:\$PATH" C-m
tmux send-keys -t "${SESSION}:0.0" "source <(docmgr _carapace bash)" C-m

tmux send-keys -t "${SESSION}:0.1" "export PATH=${DIST}:\$PATH" C-m
tmux send-keys -t "${SESSION}:0.1" "autoload -Uz compinit && compinit" C-m
tmux send-keys -t "${SESSION}:0.1" "setopt AUTO_LIST" C-m
tmux send-keys -t "${SESSION}:0.1" "zmodload zsh/complist && zstyle ':completion:*' menu select" C-m
tmux send-keys -t "${SESSION}:0.1" "source <(docmgr _carapace zsh)" C-m

sleep 0.2

# Helper to send a command line and press TAB twice, then capture last 200 lines
test_and_capture() {
  local pane="$1" line="$2" outfile="$3"
  tmux send-keys -t "${SESSION}:${pane}" "clear" C-m
  tmux send-keys -t "${SESSION}:${pane}" "${line}"
  tmux send-keys -t "${SESSION}:${pane}" C-i C-i
  sleep 0.3
  tmux capture-pane -pt "${SESSION}:${pane}" -S -200 > "${outfile}"
}

prefix="/tmp/${SESSION}"

# Bash tests
test_and_capture "0.0" "docmgr doc add --doc-type "  "${prefix}_bash_doc_type.txt"
test_and_capture "0.0" "docmgr doc add --status "    "${prefix}_bash_status.txt"
test_and_capture "0.0" "docmgr doc add --intent "    "${prefix}_bash_intent.txt"
test_and_capture "0.0" "docmgr doc add --topics "    "${prefix}_bash_topics.txt"
test_and_capture "0.0" "docmgr doc add --ticket "    "${prefix}_bash_ticket.txt"

# Zsh tests
test_and_capture "0.1" "docmgr doc add --doc-type "  "${prefix}_zsh_doc_type.txt"
test_and_capture "0.1" "docmgr doc add --status "    "${prefix}_zsh_status.txt"
test_and_capture "0.1" "docmgr doc add --intent "    "${prefix}_zsh_intent.txt"
test_and_capture "0.1" "docmgr doc add --topics "    "${prefix}_zsh_topics.txt"
test_and_capture "0.1" "docmgr doc add --ticket "    "${prefix}_zsh_ticket.txt"

echo "Captured completion outputs under /tmp using prefix: ${prefix}_*.txt"
echo "Attach with: tmux attach -t ${SESSION}"


