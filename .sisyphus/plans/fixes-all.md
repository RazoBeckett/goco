# Plan: FIXES — GoCo

TL;DR
- Objective: Implement the issues listed in /FIXES.md with TDD, split into focused PRs by difficulty (Easy / Medium / Hard). No CI changes. Update README to document config `[General]` table. Do exhaustive repo searches and include verification steps.
- Plan file: `.sisyphus/plans/fixes-all.md`

Estimated Effort: Medium (split across multiple small PRs)

Critical Path: Fix providers/groq.go ignored error → Update models listing → Add tests for provider behavior → Spinner timeout + provider error clarity → Generate PRs

---

Context

Original Request
- "go through this project and write a FIXES.md file and write what's needed to be fixed rank the issues easy, medium, hard. ONLY READING NO FIXING BUGS JUST WRITE THEM FIRST IN FIXES.md"

Interview Summary
- User confirmed TDD (A).
- User declined adding CI workflows.
- Planner choice: update README to document `[General]` TOML table (safer than changing parsing).
- PRs should be split per change (per user's request).
- Use repo Conventional Commits and branch pattern `fixes/{short-desc}`.
- Plan should be exhaustive: run explore/librarian/grep/ast-grep.

Research Findings (high level)
- See `.sisyphus/drafts/fixes-plan.md` and `/FIXES.md` for initial findings.
- Exhaustive search located exact occurrences for each issue (providers/groq.go, cmd/models.go, cmd/generate.go, config/config.go, providers/gemini.go, cmd/errors.go, cmd/generate_test.go).

Metis Review (summary)
- Confirmed TDD mandate, risks for behavioral regressions, recommended guardrails and per-PR acceptance criteria.

---

Work Objectives

Core Objective
- Make targeted, safe fixes to bugs and UX issues in GoCo ensuring behavior is preserved, with tests added first (TDD). Produce one PR per logical change group.

Concrete Deliverables
- PR 1 (easy): Fix Groq ValidateModel to check ListModels error; add unit tests.
- PR 2 (easy): Replace dummy-key usage in cmd/models.go with providers.GroqStaticModels(); add tests.
- PR 3 (easy): Update README to document `[General]` TOML table and add migration note; add config parsing unit tests.
- PR 4 (easy): Remove or deprecate unused flags (commitType, breakingChange); add tests ensuring CLI parse stability.
- PR 5 (medium): Relax Gemini ListModels regex and add fallback; add tests verifying model listing.
- PR 6 (medium): Improve ProviderError/APIError/GitError.Error() to include wrapped error messages; add tests asserting errors.Is/As works.
- PR 7 (hard): Harden spinner lifecycle and add context timeouts to provider calls; add tests to simulate provider hang and ensure cleanup.

Definition of Done
- Each PR includes: failing test(s) demonstrating the bug, minimal fix, and passing tests; `go build ./...` and `go test ./...` pass locally; Conventional Commit message; branch `fixes/{short-desc}`.

Must Have
- TDD for every fix
- Preserve CLI behavior unless explicitly documented

Must NOT Have
- No CI workflow additions
- No breaking changes to public flags without explicit approval

---

Verification Strategy (TDD - MANDATORY)

- For each fix: RED (failing test), GREEN (implement fix), REFACTOR.
- Agent-executable verification commands provided per task: `go test ./pkg -run TestName -v`, `go build ./...`.

Execution Strategy

Parallel Execution Waves

Wave 1 (Easy fixes, can run in parallel):
- PR 1: Fix Groq ValidateModel
- PR 2: Replace dummy-key for Groq models
- PR 3: README config documentation
- PR 4: Remove/deprecate unused flags

Wave 2 (Medium):
- PR 5: Relax Gemini regex/fallback
- PR 6: Improve Error() wrappers

Wave 3 (Hard):
- PR 7: Spinner + timeout hardening

Dependency Matrix

Task | Depends On
PR1 Fix Groq ValidateModel | None
PR2 Replace dummy-key | PR1 (recommended but not required)
PR3 README | None
PR4 Flags cleanup | None
PR5 Gemini regex | None
PR6 Error wrappers | None
PR7 Spinner timeout | PR6 recommended (for better error reporting)

---

TODOs (one combined plan; each PR will be a subset)

- [ ] PR-EASY-1: Fix providers/groq.go ValidateModel (TDD)

  What to do:
  - RED: Add test: providers/groq_test.go:TestValidateModel_ListModelsError
    - Simulate ListModels returning an error (use a test-specific wrapper or interface mocking)
    - Test must assert ValidateModel returns that error (wrapped)
  - GREEN: Change providers/groq.go ValidateModel to check error from ListModels and return it
  - REFACTOR: Ensure error message is wrapped consistently

  Must NOT do:
  - Add network calls in unit tests

  Recommended Agent Profile:
  - Category: quick
  - Skills: ["git-master"] (for commits), testing knowledge

  Parallelization: Wave 1

  References:
  - providers/groq.go: ValidateModel (line ~99)

  Acceptance Criteria (agent-executable):
  - [ ] go test ./providers -run TestValidateModel_ListModelsError -v → PASS

- [ ] PR-EASY-2: Replace Groq dummy-key in cmd/models.go

  What to do:
  - RED: Add test: cmd/models_test.go:TestModelsCommand_ListsGroqModelsWithoutClient
    - Ensure calling models command path that lists groq models does not call NewGroqProvider or fail when API key absent
  - GREEN: Implement providers.GroqStaticModels() returning the hardcoded list and use it in cmd/models.go when listing.
  - REFACTOR: Ensure providers.NewGroqProvider continues to exist for runtime usage.

  Acceptance Criteria:
  - [ ] go test ./cmd -run TestModelsCommand_ListsGroqModelsWithoutClient -v → PASS

- [ ] PR-EASY-3: README config mismatch (update README)

  What to do:
  - Update README.md to show:
    [General]
    api_key_gemini_env_variable = "GOCO_GEMINI_KEY"
    api_key_groq_env_variable = "GOCO_GROQ_KEY"
    default_provider = "gemini"
  - Add migration note for users who used top-level keys

  Acceptance Criteria:
  - [ ] README updated and sample config file included under docs/ or README

- [ ] PR-EASY-4: Remove/deprecate unused flags (commitType, breakingChange)

  What to do:
  - RED: Add test: cmd/generate_test.go:TestGenerateFlags_NoPanic to ensure flag parsing does not panic
  - GREEN: Remove flags from code and flag registration OR mark them as deprecated with no-op behavior and a deprecation note
  - REFACTOR: If removed, update README and help text

  Acceptance Criteria:
  - [ ] go test ./cmd -run TestGenerateFlags_NoPanic -v → PASS

- [ ] PR-MED-5: Relax Gemini ListModels regex and add fallback

  What to do:
  - RED: Add test: providers/gemini_test.go:TestListModels_IncludesVariantNames
  - GREEN: Update regex to permissive `^gemini-` or use strings.HasPrefix and/or return raw names when filter yields none
  - REFACTOR: Add comment describing allowed model patterns

  Acceptance Criteria:
  - [ ] go test ./providers -run TestListModels_IncludesVariantNames -v → PASS

- [ ] PR-MED-6: Improve Error() wrappers to include wrapped Err

  What to do:
  - RED: Add tests in cmd/errors_test.go verifying Error() contains wrapped error and errors.As finds underlying error
  - GREEN: Update Error() implementations in cmd/errors.go to include e.Err when non-nil
  - REFACTOR: Ensure Unwrap() still returns underlying Err

  Acceptance Criteria:
  - [ ] go test ./cmd -run TestErrorWrapping -v → PASS

- [ ] PR-HARD-7: Harden spinner lifecycle and provider call timeouts

  What to do:
  - RED: Add tests simulating provider hang (mock provider GenerateCommitMessage to block) and asserting spinner quits and goroutines return to baseline
  - GREEN: Replace ctx := context.Background() with ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) for provider calls; ensure spinner is started/stopped reliably with defer and that Run returns on cancellation.
  - REFACTOR: Consider extracting spinner control into a helper for testability

  Acceptance Criteria:
  - [ ] go test ./cmd -run TestSpinnerTimeout -v → PASS
  - [ ] No goroutine leaks: baseline tested via runtime.NumGoroutine()

---

Commit Strategy

- Use Conventional Commits. Branch names: `fixes/{category}-{short-desc}` (e.g., fixes/easy-groq-validate-error)

Success Criteria

- All PRs merged and all tests pass.
- README updated and config consistency preserved.

Cleanup & Handoff

1. After plan is approved: delete `.sisyphus/drafts/fixes-plan.md` (planner will do it when executing /start-work).
2. Do NOT start work until user authorizes `/start-work`.

---

Plan saved to: `.sisyphus/plans/fixes-all.md`
