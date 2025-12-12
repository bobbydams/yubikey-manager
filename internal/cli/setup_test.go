package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSetupCmd(t *testing.T) {
	cmd := newSetupCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "setup", cmd.Use)
}

func TestNewSetupBatchCmd(t *testing.T) {
	cmd := newSetupBatchCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "setup-batch", cmd.Use)
}
