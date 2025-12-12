package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVerifyCmd(t *testing.T) {
	cmd := newVerifyCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "verify", cmd.Use)
	assert.Contains(t, cmd.Aliases, "check")
}

func TestGetGitConfig(t *testing.T) {
	// Test that getGitConfig can be called
	// It executes git config command, so we can't easily test the output
	// without mocking, but we can verify it doesn't panic
	result := getGitConfig("user.name")
	// Result will be empty or contain git config value
	_ = result
	assert.NotPanics(t, func() {
		getGitConfig("user.email")
		getGitConfig("commit.gpgsign")
		getGitConfig("user.signingkey")
	})
}
