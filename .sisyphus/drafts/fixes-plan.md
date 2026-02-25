# Draft: FIXES — GoCo

## Requirements (confirmed)
- Create a single, comprehensive work plan to implement the issues listed in `/FIXES.md`.
- The plan must cover all items in FIXES.md (Easy / Medium / Hard) in one `.sisyphus/plans/*.md` file when generation is triggered.
- The plan should include concrete TODOs, acceptance criteria (agent-executable where possible), recommended agent profiles, references to code locations, and test/CI changes as needed.

## Technical Decisions (pending)
- Test strategy: NOT DECIDED (see open questions)
- CI changes: NOT DECIDED (see open questions)
- README vs code fix approach: NOT DECIDED (change README OR change config struct)

## Research Findings (from quick repo scan)
- Repo module: `github.com/razobeckett/goco` (go.mod)
- Main packages: `cmd`, `providers`, `config`
- Config logic in `config/config.go` uses a `[General]` TOML table; README shows top-level keys (mismatch)
- Tests: `cmd/generate_test.go` exists and uses `git` binaries (integration-style unit test)
- Providers: `providers/gemini.go` (google.golang.org/genai) and `providers/groq.go` (github.com/conneroisu/groq-go)
- FIXES.md issues enumerated and cover: ignored errors, UX prompting, regex filtering, spinner/concurrency, error wrapping, CI absence

## Open Questions (decisions recorded)
1. Test strategy — Decision: A (TDD) — user confirmed TDD for all fixes.

2. CI workflow changes — Decision: NO CI — user opted out of adding CI workflow.

3. README vs config: Decision: Planner chose to update README to match the existing code (document `[General]` table). Rationale: safer, preserves runtime behaviour and avoids parsing changes.

4. Priority & timebox: Decision: Split PRs per user's request. Each PR will contain a cohesive set of changes (Easy / Medium / Hard groups). Branch naming: `fixes/{category}-{short-desc}`.

5. Commit/branch conventions: Decision: Use Conventional Commits and branch pattern `fixes/{short-description}` (confirmed).

## Scope Boundaries
INCLUDE:
- All items listed in `/FIXES.md` (items 1–11).
- Adding tests where missing or required by chosen test strategy.
- Adding/modifying CI workflow if approved.

EXCLUDE:
- Any new features beyond the listed fixes.
- Large refactors not requested in FIXES.md (unless required to safely implement a fix; document justification).

## Next actions (what I'll do once you answer the open questions)
1. Lock test strategy and CI decision.
2. Generate a single comprehensive plan file: `.sisyphus/plans/fixes-all.md` containing TL;DR, objectives, verification strategy, execution waves, and a fully detailed TODO section covering every issue.
3. Run Metis review (pre-plan validation) and incorporate findings.
4. Present plan summary and request approval to create PR(s).

---

Recorded by Prometheus at: 2026-01-30
