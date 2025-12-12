package executor

import (
	"context"
)

// MockExecutor implements Executor for testing purposes.
// It allows pre-defining command outputs and tracking command invocations.
type MockExecutor struct {
	// Outputs maps command+args to expected output
	Outputs map[string][]byte
	// Errors maps command+args to expected error
	Errors map[string]error
	// Calls tracks all command invocations for verification
	Calls []CommandCall
	// InteractiveCalls tracks interactive command invocations
	InteractiveCalls []CommandCall
}

// CommandCall represents a single command invocation.
type CommandCall struct {
	Name string
	Args []string
}

// NewMockExecutor creates a new MockExecutor instance.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		Outputs:          make(map[string][]byte),
		Errors:           make(map[string]error),
		Calls:            make([]CommandCall, 0),
		InteractiveCalls: make([]CommandCall, 0),
	}
}

// SetOutput sets the expected output for a command.
// The key should be in the format "command arg1 arg2 ..."
func (m *MockExecutor) SetOutput(key string, output []byte) {
	m.Outputs[key] = output
}

// SetError sets the expected error for a command.
func (m *MockExecutor) SetError(key string, err error) {
	m.Errors[key] = err
}

// Run executes a mocked command and returns the predefined output or error.
func (m *MockExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	key := m.buildKey(name, args...)
	m.Calls = append(m.Calls, CommandCall{Name: name, Args: args})

	if err, ok := m.Errors[key]; ok {
		return nil, err
	}

	if output, ok := m.Outputs[key]; ok {
		return output, nil
	}

	// If no output is set, return empty output
	return []byte{}, nil
}

// RunInteractive executes a mocked interactive command.
func (m *MockExecutor) RunInteractive(ctx context.Context, name string, args ...string) error {
	key := m.buildKey(name, args...)
	m.InteractiveCalls = append(m.InteractiveCalls, CommandCall{Name: name, Args: args})

	if err, ok := m.Errors[key]; ok {
		return err
	}

	return nil
}

// buildKey creates a key from command name and args for lookup.
func (m *MockExecutor) buildKey(name string, args ...string) string {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	return key
}

// Reset clears all recorded calls and outputs.
func (m *MockExecutor) Reset() {
	m.Calls = make([]CommandCall, 0)
	m.InteractiveCalls = make([]CommandCall, 0)
	m.Outputs = make(map[string][]byte)
	m.Errors = make(map[string]error)
}

// VerifyCall checks if a specific command was called.
func (m *MockExecutor) VerifyCall(name string, args ...string) bool {
	expected := CommandCall{Name: name, Args: args}
	for _, call := range m.Calls {
		if call.Name == expected.Name && m.argsEqual(call.Args, expected.Args) {
			return true
		}
	}
	return false
}

// argsEqual compares two argument slices.
func (m *MockExecutor) argsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
