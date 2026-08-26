## Agent skills

### Issue tracker

Issues live as GitHub issues in `thomaslgrega/bitelyapi`, managed via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical triage roles, each label string equal to its name. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

## Commits

Commit only when the user asks for it in that turn. Otherwise leave the work uncommitted for them to review.

## Comments

A comment earns its line by carrying what the code cannot: a constraint the
compiler does not express, an alternative tried and rejected, the spec paragraph
that forces this shape. Needing a comment to explain *what* the code does is a
sign it wants a better name or a smaller function.

Before writing one, ask what a reader loses if it is not there. If nothing, it
was restating the code.

- **One line is the default.** Running long usually means padded wording or a
  function doing too much. Fix that rather than relocating the prose into a new
  doc.
- **Present tense, describing the code as it stands.** Git holds the history, so
  renames, old bugs and commit SHAs stay out.
- **Architectural decisions land in an ADR first**, then get cited by number.
  This repo owns the shared definitions, so a cited ADR is the one place they
  live — a paraphrase alongside the citation drifts.

## TDD

Every change goes red → green, and the red step is not optional. Skipping straight to implementation is the one failure mode this rule exists to stop.

1. **Red.** Write tests that capture the desired behaviour, run them, and confirm they fail for the reason you expect. For a bug, the failing test reproduces the bug.
2. **Green.** Write the implementation that makes those tests pass.
3. **Verify.** Run `go test ./...` and `go vet ./...` and confirm both are clean before reporting the work done.

Report step 3's actual output. A suite you did not run is not green.
