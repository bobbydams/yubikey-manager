package executor

import (
	"context"
	"fmt"
	"os"
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
			// Include stderr in the error message for better diagnostics
			stderr := string(exitErr.Stderr)
			if stderr != "" {
				return output, fmt.Errorf("command failed with exit code %d: %s: %w", exitErr.ExitCode(), stderr, err)
			}
			return output, fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return output, fmt.Errorf("failed to execute command: %w", err)
	}
	return output, nil
}

// RunInteractive executes a command with interactive I/O.
func (e *RealExecutor) RunInteractive(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	// Connect to the terminal for interactive I/O
	// This is essential for pinentry to work correctly
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set GPG_TTY for pinentry to find the terminal
	// This is required for terminal-based pinentry programs
	if tty, err := os.Readlink("/dev/fd/0"); err == nil {
		cmd.Env = append(os.Environ(), "GPG_TTY="+tty)
	} else {
		// Fallback: try to get TTY from environment or use /dev/tty
		if tty := os.Getenv("GPG_TTY"); tty != "" {
			cmd.Env = append(os.Environ(), "GPG_TTY="+tty)
		} else {
			cmd.Env = append(os.Environ(), "GPG_TTY=/dev/tty")
		}
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// GPG's "save" command returns exit code 2 when there are no changes to save.
			// This is a success case, not an error. For example, when a key is already
			// saved or when "save" is called but no changes were made.
			if name == "gpg" && exitCode == 2 {
				return nil
			}
			return fmt.Errorf("command failed with exit code %d: %w", exitCode, err)
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}
