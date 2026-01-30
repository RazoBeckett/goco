package cmd

import (
	"errors"
	"fmt"
)

type ValidationError struct {
	Field   string
	Message string
	Help    string
}

func (e *ValidationError) Error() string {
	if e.Help != "" {
		return fmt.Sprintf("%s: %s\n\nHelp: %s", e.Field, e.Message, e.Help)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error (%s): %s", e.Field, e.Message)
}

type ProviderError struct {
	Provider string
	Message  string
	Err      error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider error (%s): %s", e.Provider, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

type GitError struct {
	Command string
	Message string
	Err     error
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git error: %s failed: %s", e.Command, e.Message)
}

func (e *GitError) Unwrap() error {
	return e.Err
}

type APIError struct {
	Message string
	Err     error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("API error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("API error: %s", e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

var (
	ErrNoEditor      = errors.New("no text editor available")
	ErrGitRepository = errors.New("not a git repository")
	ErrNoStagedFiles = errors.New("no staged files found")
)
