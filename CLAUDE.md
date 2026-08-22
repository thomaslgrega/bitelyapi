## Agent skills

### Issue tracker

Issues live as GitHub issues in `thomaslgrega/bitelyapi`, managed via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical triage roles, each label string equal to its name. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

## Commits

Commit only when the user asks for it in that turn. Otherwise leave the work uncommitted for them to review.

## TDD

Every change goes red → green, and the red step is not optional. Skipping straight to implementation is the one failure mode this rule exists to stop.

1. **Red.** Write tests that capture the desired behaviour, run them, and confirm they fail for the reason you expect. For a bug, the failing test reproduces the bug.
2. **Green.** Write the implementation that makes those tests pass.
3. **Verify.** Run `go test ./...` and `go vet ./...` and confirm both are clean before reporting the work done.

Report step 3's actual output. A suite you did not run is not green.
