package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockExecutor_Run(t *testing.T) {
	mock := NewMockExecutor()

	// Set up expected output
	key := "gpg --list-keys"
	expectedOutput := []byte("test output")
	mock.SetOutput(key, expectedOutput)

	output, err := mock.Run(context.Background(), "gpg", "--list-keys")

	require.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
	assert.True(t, mock.VerifyCall("gpg", "--list-keys"))
}

func TestMockExecutor_Run_WithError(t *testing.T) {
	mock := NewMockExecutor()

	key := "gpg --invalid"
	expectedError := assert.AnError
	mock.SetError(key, expectedError)

	output, err := mock.Run(context.Background(), "gpg", "--invalid")

	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestMockExecutor_RunInteractive(t *testing.T) {
	mock := NewMockExecutor()

	err := mock.RunInteractive(context.Background(), "gpg", "--edit-key", "123")

	require.NoError(t, err)
	assert.Len(t, mock.InteractiveCalls, 1)
	assert.Equal(t, "gpg", mock.InteractiveCalls[0].Name)
}

func TestMockExecutor_Reset(t *testing.T) {
	mock := NewMockExecutor()

	mock.SetOutput("test", []byte("output"))
	_, err := mock.Run(context.Background(), "test")
	assert.NoError(t, err)

	mock.Reset()

	assert.Len(t, mock.Calls, 0)
	assert.Len(t, mock.Outputs, 0)
	assert.Len(t, mock.Errors, 0)
}
