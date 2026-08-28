#!/bin/bash
# PostToolUse hook: hand a freshly-created PR to Codex for review.
# Hook mode reads the payload on stdin, claims the PR, and detaches a worker.
# Worker mode (--run) runs codex and records the review only once it lands.
set -uo pipefail

export PATH="$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH"
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$PWD}"
LOG_DIR="$PROJECT_DIR/.claude/hooks/logs"
# Injected so a test can substitute stubs for the real reviewer and CLI.
CODEX_BIN="${CODEX_BIN:-codex}"
GH_BIN="${GH_BIN:-gh}"

run_review() {
  local pr=$1 repo=$2 status sentinel
  # Stamped into the comment so delivery is confirmed by this run's own mark and
  # not by an unrelated bot or human commenting while codex works.
  sentinel="codex-review-$pr-$(date +%s)-$$"

  "$CODEX_BIN" exec \
    -C "$PROJECT_DIR" \
    -s workspace-write \
    -c sandbox_workspace_write.network_access=true \
    "/code-review PR #$pr of $repo.

Establish the change set without touching the working tree — do not check out, switch, stash, or reset anything:

  gh pr view $pr --repo $repo --json baseRefName,baseRefOid,headRefOid,title,body
  git fetch origin pull/$pr/head
  git diff <baseRefOid>...<headRefOid>

Use <baseRefOid> as the fixed point.

Posting on the PR is the deliverable: a review that ends in your final response has not been delivered. Write the body to a file outside the repository, under \$TMPDIR, so the working tree stays clean, then post exactly one comment:

  gh pr comment $pr --repo $repo --body-file <that file>

End the body with the line <!-- $sentinel --> exactly as written; it is how this run confirms its comment landed.

Lead with a one-line verdict, then the findings ordered most severe first, each naming file and line. If nothing is worth acting on, post that verdict rather than posting nothing. Do not commit, push, edit repository files, approve, request changes, or merge."
  status=$?

  # Suppress future reviews only once this run's comment is on the PR, so a
  # failed or unconfirmed run is retried rather than silently swallowed.
  if [ "$status" -eq 0 ] && "$GH_BIN" pr view "$pr" --repo "$repo" --json comments \
      -q '.comments[].body' 2>/dev/null | grep -qF "$sentinel"; then
    : > "$LOG_DIR/.reviewed-$pr"
  fi
  rmdir "$LOG_DIR/.inflight-$pr" 2>/dev/null
}

if [ "${1:-}" = "--run" ]; then
  run_review "$2" "$3"
  exit 0
fi

mkdir -p "$LOG_DIR"
payload=$(cat)

repo=$(cd "$PROJECT_DIR" && "$GH_BIN" repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null)
[ -z "$repo" ] && exit 0

# Only the tool result names the PR just created — tool_input may quote others.
# `gh pr create` prints the URL alone on a line, so requiring a whole-line match
# rejects PR URLs a command merely mentions in prose.
response=$(printf '%s' "$payload" | jq -r '.tool_response // empty' 2>/dev/null)
repo_re=$(printf '%s' "$repo" | sed 's/[][\.*^$+?(){}|\\]/\\&/g')
pr=$(printf '%s' "$response" | jq -r '.stdout // empty' 2>/dev/null \
  | grep -Eo "^https://github\.com/$repo_re/pull/[0-9]+$" | tail -1 | grep -oE '[0-9]+$')
# The MCP tool takes owner/repo, so a session rooted here can open a PR
# elsewhere. Its number is only ours when those fields name this repository.
if [ -z "$pr" ] && [ "$(printf '%s' "$payload" | jq -r '"\(.tool_input.owner)/\(.tool_input.repo)"' 2>/dev/null)" = "$repo" ]; then
  pr=$(printf '%s' "$response" | jq -r '.number // empty' 2>/dev/null)
fi
[ -z "$pr" ] && exit 0

# Release a claim whose worker never started or was killed before it could.
find "$LOG_DIR" -maxdepth 1 -type d -name '.inflight-*' -mmin +120 -exec rmdir {} \; 2>/dev/null

# mkdir is atomic, so overlapping hook processes cannot both claim the PR.
mkdir "$LOG_DIR/.inflight-$pr" 2>/dev/null || exit 0
if [ -e "$LOG_DIR/.reviewed-$pr" ]; then
  rmdir "$LOG_DIR/.inflight-$pr"
  exit 0
fi

# Invoked through bash: the hook itself is run as `bash <script>`, so the file
# is not required to carry the exec bit.
nohup bash "$0" --run "$pr" "$repo" >"$LOG_DIR/pr-$pr.log" 2>&1 &

echo "{\"systemMessage\": \"Codex review of PR #$pr started (log: .claude/hooks/logs/pr-$pr.log)\"}"
