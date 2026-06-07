package ai

import "fmt"

const conventionalCommitsSpec = `
Conventional Commits specification:

The commit message MUST be structured as:
  <type>[optional scope]: <description>
  [blank line]
  [optional body]

Types MUST be one of:
  feat     — a new feature
  fix      — a bug fix
  docs     — documentation only changes
  style    — formatting, missing semi-colons, etc; no code change
  refactor — a code change that neither fixes a bug nor adds a feature
  perf     — a code change that improves performance
  test     — adding missing tests or correcting existing tests
  chore    — changes to the build process or auxiliary tools
  ci       — changes to CI configuration files and scripts
  build    — changes that affect the build system or external dependencies

Rules:
  - type and description are mandatory
  - scope is optional and MUST be in parentheses after the type
  - description MUST start with a lowercase letter
  - description MUST NOT end with a period
  - subject line (type + scope + description) MUST be <= 72 characters
  - body is optional, separated from subject by a blank line
  - breaking changes MUST append ! before the colon, e.g. feat!: drop support
  - breaking changes MAY include BREAKING CHANGE: footer in the body
`

func buildPrompt(gitStatus, gitDiff, customInstructions string) string {
	prompt := fmt.Sprintf(
		"Generate a Conventional Commit based strictly on the following:\n\n"+
			"Git Status:\n%s\n\n"+
			"Git Diff:\n%s\n\n"+
			"%s\n"+
			"Before responding, you MUST:\n"+
			"- ONLY output the commit message and description.\n"+
			"- There must be a commit summary (one line) at the top, then an empty line, then the commit description below.\n"+
			"- DO NOT include markdown, code blocks, quotes, or any formatting.\n"+
			"- Output MUST be plain text only.\n"+
			"- Do not add extra explanations, notes, or commentary.\n"+
			"- The first line is the commit summary, the rest is the description.\n"+
			"- Follow the specification above exactly.\n"+
			"- No extra lines before or after the commit message.\n",
		gitStatus,
		gitDiff,
		conventionalCommitsSpec,
	)

	if customInstructions != "" {
		prompt += fmt.Sprintf("\nAdditional Instructions:\n%s\n", customInstructions)
	}

	return prompt
}
