package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRevokeCmd(t *testing.T) {
	cmd := newRevokeCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "revoke", cmd.Use)
}
