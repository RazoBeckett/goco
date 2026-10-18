package ai

import "strings"

// systemPrompt carries the Conventional Commits rules. It is stable across
// calls so providers that cache prompts can reuse it.
const systemPrompt = `You write Conventional Commit messages from a git diff.

Return only the commit message. Do not use markdown, code fences, quotes, or commentary.

Format the subject line as:
  <type>[optional scope]: <description>

Rules:
  - type must be one of: feat, fix, docs, style, refactor, perf, test, chore, ci, build
  - scope is optional and goes in parentheses right after the type
  - the description starts with a lowercase letter and does not end with a period
  - the whole subject line is 72 characters or fewer
  - put a blank line after the subject, then an optional body when the change needs explanation
  - mark a breaking change with ! before the colon, for example: feat!: drop the v1 API`

// buildUserPrompt assembles the change description sent to the model.
func buildUserPrompt(gitStatus, gitDiff, customInstructions, recentLog string) string {
	var b strings.Builder

	b.WriteString("Git Status:\n")
	b.WriteString(gitStatus)
	b.WriteString("\n\nGit Diff:\n")
	b.WriteString(gitDiff)

	if strings.TrimSpace(recentLog) != "" {
		b.WriteString("\n\nRecent Commits (for context, do not repeat them):\n")
		b.WriteString(recentLog)
	}

	if customInstructions != "" {
		b.WriteString("\n\nAdditional Instructions:\n")
		b.WriteString(customInstructions)
	}

	return b.String()
}
