#!/bin/bash
# Regression coverage for codex-pr-review.sh. Stubs `gh` and `codex` through the
# script's GH_BIN/CODEX_BIN injection points, so no network or real PR is touched.
set -uo pipefail

HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/codex-pr-review.sh"
REPO="thomaslgrega/bitelyapi"
failures=0

pass() { printf 'ok   - %s\n' "$1"; }
fail() { printf 'FAIL - %s\n     %s\n' "$1" "$2"; failures=$((failures + 1)); }

check() { # name expected actual
  if [ "$2" = "$3" ]; then pass "$1"; else fail "$1" "expected [$2], got [$3]"; fi
}

setup() {
  WORK=$(mktemp -d)
  export CLAUDE_PROJECT_DIR="$WORK/project"
  LOGS="$CLAUDE_PROJECT_DIR/.claude/hooks/logs"
  mkdir -p "$CLAUDE_PROJECT_DIR"

  # Stub gh: serves the repo name, and a comment list the test controls.
  export GH_BIN="$WORK/gh"
  : > "$WORK/comments"
  cat > "$GH_BIN" <<EOF
#!/bin/bash
printf '%s\n' "\$*" >> "$WORK/gh-calls"
case "\$*" in
  *"repo view"*)  echo "$REPO" ;;
  *"pr view"*)    cat "$WORK/comments" ;;
esac
EOF
  chmod +x "$GH_BIN"

  # Stub codex: records argv, and posts the sentinel only when told to.
  export CODEX_BIN="$WORK/codex"
  cat > "$CODEX_BIN" <<EOF
#!/bin/bash
printf '%s\n' "\$@" > "$WORK/argv"
if [ -f "$WORK/codex-posts" ]; then
  grep -o 'codex-review-[0-9a-z.-]*' <<< "\$*" | tail -1 >> "$WORK/comments"
fi
exit \$(cat "$WORK/codex-status" 2>/dev/null || echo 0)
EOF
  chmod +x "$CODEX_BIN"
  rm -f "$WORK/codex-posts"
}

teardown() { rm -rf "$WORK"; }

# Runs the hook in hook mode and waits for the detached worker to settle.
run_hook() {
  printf '%s' "$1" | bash "$HOOK"
  local waited=0
  while [ -d "$LOGS"/.inflight-* ] 2>/dev/null && [ $waited -lt 100 ]; do
    sleep 0.1; waited=$((waited + 1))
  done
  sleep 0.3
}

create_payload() { # pr
  jq -n --arg url "https://github.com/$REPO/pull/$1" \
    '{tool_name:"Bash",tool_input:{command:"gh pr create"},tool_response:{stdout:($url + "\n")}}'
}

mcp_payload() { # pr owner repo
  jq -n --argjson n "$1" --arg o "$2" --arg r "$3" \
    '{tool_name:"mcp__github__create_pull_request",tool_input:{owner:$o,repo:$r},tool_response:{number:$n}}'
}

# --- a created PR reaches codex ------------------------------------------------
setup
run_hook "$(create_payload 42)"
check "gh pr create launches a review of the right PR" \
  "yes" "$(grep -qF "/code-review PR #42 of $REPO." "$WORK/argv" && echo yes || echo no)"
check "review runs in the project directory" \
  "yes" "$(grep -qxF "$CLAUDE_PROJECT_DIR" "$WORK/argv" && echo yes || echo no)"
teardown

# --- payloads that must not launch a review ------------------------------------
for desc_payload in \
  "ordinary command output|$(jq -n '{tool_response:{stdout:"all tests passed\n"}}')" \
  "a PR URL mentioned in prose|$(jq -n --arg u "see https://github.com/$REPO/pull/99 for context" '{tool_response:{stdout:($u + "\n")}}')" \
  "a PR URL for another repository|$(jq -n '{tool_response:{stdout:"https://github.com/thomaslgrega/bitely-ios/pull/50\n"}}')" \
  "an MCP PR opened in another repository|$(mcp_payload 50 thomaslgrega bitely-ios)"
do
  setup
  run_hook "${desc_payload#*|}"
  check "ignores ${desc_payload%%|*}" "no" "$([ -f "$WORK/argv" ] && echo yes || echo no)"
  teardown
done

# --- MCP fallback --------------------------------------------------------------
setup
run_hook "$(mcp_payload 7 thomaslgrega bitelyapi)"
check "accepts an MCP PR opened in this repository" \
  "yes" "$(grep -qF "/code-review PR #7 of $REPO." "$WORK/argv" && echo yes || echo no)"
teardown

# --- the reviewed marker tracks this run's own comment -------------------------
setup
touch "$WORK/codex-posts"
run_hook "$(create_payload 42)"
check "marks reviewed once this run's comment lands" \
  "yes" "$([ -f "$LOGS/.reviewed-42" ] && echo yes || echo no)"
teardown

setup
echo "codex-review-42-someone-elses-run" > "$WORK/comments"
run_hook "$(create_payload 42)"
check "does not mark reviewed when only an unrelated comment appears" \
  "no" "$([ -f "$LOGS/.reviewed-42" ] && echo yes || echo no)"
teardown

setup
touch "$WORK/codex-posts"; echo 1 > "$WORK/codex-status"
run_hook "$(create_payload 42)"
check "does not mark reviewed when codex fails" \
  "no" "$([ -f "$LOGS/.reviewed-42" ] && echo yes || echo no)"
teardown

# --- claims ---------------------------------------------------------------------
setup
mkdir -p "$LOGS"; : > "$LOGS/.reviewed-42"
run_hook "$(create_payload 42)"
check "skips a PR already reviewed" "no" "$([ -f "$WORK/argv" ] && echo yes || echo no)"
teardown

setup
mkdir -p "$LOGS/.inflight-42"
printf '%s' "$(create_payload 42)" | bash "$HOOK"
sleep 0.3
check "skips a PR another worker holds" "no" "$([ -f "$WORK/argv" ] && echo yes || echo no)"
teardown

setup
touch "$WORK/codex-posts"
run_hook "$(create_payload 42)"
check "releases the claim when the worker finishes" \
  "no" "$([ -d "$LOGS/.inflight-42" ] && echo yes || echo no)"
teardown

if [ "$failures" -gt 0 ]; then
  printf '\n%d test(s) failed\n' "$failures"
  exit 1
fi
printf '\nall tests passed\n'
