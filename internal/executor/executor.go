package executor

import (
	"context"
	"fmt"
	"os/exec"
)

// Executor provides an interface for executing external commands.
// This abstraction allows for easy mocking in tests.
type Executor interface {
	// Run executes a command and returns its stdout output.
	// Returns an error if the command fails or if stdout cannot be captured.
	Run(ctx context.Context, name string, args ...string) ([]byte, error)

	// RunInteractive executes a command with interactive I/O (stdin/stdout/stderr).
	// This is used for commands that require user interaction like gpg --edit-key.
	RunInteractive(ctx context.Context, name string, args ...string) error
}

// RealExecutor implements Executor using the os/exec package.
type RealExecutor struct{}

// NewRealExecutor creates a new RealExecutor instance.
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Run executes a command and returns its stdout output.
func (e *RealExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return output, fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return output, fmt.Errorf("failed to execute command: %w", err)
	}
	return output, nil
}

// RunInteractive executes a command with interactive I/O.
func (e *RealExecutor) RunInteractive(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}
