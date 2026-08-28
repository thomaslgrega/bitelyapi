---
name: code-review
description: Review a change set for correctness, security, performance, test coverage, and repository-standard violations. Use for pull requests, branches, or working-tree changes when the user asks for a technical code review or pre-commit quality gate.
---

# Code Review

Review changed code against a fixed point and report only actionable findings.

## Establish scope and standards

Use the fixed point supplied by the caller. For a pull request, prefer its base SHA or base branch. If no fixed point is supplied, review tracked and untracked working-tree changes against `HEAD`.

Read every applicable `AGENTS.md` from the repository root to each changed file. Follow its pointers when the changed area triggers them. Inspect `README.md`, `CONTEXT.md`, ADRs, and other repository documentation only when they define behavior or standards relevant to the change.

Gather the change set with the appropriate commands:

```bash
git status --short
git diff --stat <fixed-point>...HEAD
git diff --name-status <fixed-point>...HEAD
git diff <fixed-point>...HEAD
```

For working-tree changes, use:

```bash
git status --short
git diff --stat HEAD
git diff HEAD
git ls-files --others --exclude-standard
```

Read every changed or new file in full. For deleted files, inspect the version at the fixed point when needed to understand the effect of removal. Include generated files only when they carry behavior that is not reliably established by their source.

## Review

Trace changed behavior through callers, state transitions, persistence, network boundaries, and tests. Look for:

1. **Correctness**
   - Incorrect conditions, boundary errors, invalid state transitions, and broken invariants
   - Missing error handling, cancellation, cleanup, or concurrency protection
   - Contract mismatches between callers, APIs, storage, and UI
2. **Security and privacy**
   - Injection, authorization, secret exposure, unsafe parsing, and untrusted input handling
   - Sensitive data stored, logged, or transmitted beyond its intended boundary
3. **Performance and reliability**
   - Avoidable repeated work, unbounded resource use, leaks, races, and failure amplification
4. **Tests**
   - Behavior changes without meaningful coverage
   - Tests that cannot detect the regression they claim to cover
   - Shared mutable fixtures or other sources of order dependence
5. **Repository standards**
   - Violations of applicable `AGENTS.md`, referenced specifications, ADRs, formatting, or type-checking rules

Focus on defects introduced by the change. Treat style preferences and speculative improvements as non-findings unless they cause a concrete maintenance or correctness risk.

## Verify findings

Prove each finding from the code path and surrounding context. Run focused tests or read-only checks when the environment supports them and they materially increase confidence. If a required check cannot run, state that limitation without presenting an unverified suspicion as fact.

Before reporting a finding, confirm all of the following:

- The changed lines introduce or expose the problem.
- A concrete input, state, or execution path triggers it.
- The impact is material enough that the author would likely fix it.
- The cited line is the narrowest useful location for the review comment.

## Report

List findings first, ordered by severity. Use this shape for each finding:

```text
severity: critical|high|medium|low
file: path/to/file
line: 42
issue: One-line description
detail: Trigger, impact, and evidence.
suggestion: Smallest viable direction for a fix.
```

Then include compact statistics for files added, modified, deleted, lines added, and lines deleted. If no findings survive verification, write:

```text
Code review passed. No technical issues detected.
```

Return the report in the final response. Save it under `.agents/code-reviews/` only when the user explicitly requests a file. Do not modify the reviewed code during a review.
