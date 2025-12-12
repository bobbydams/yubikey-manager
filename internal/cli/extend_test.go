package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExtendCmd(t *testing.T) {
	cmd := newExtendCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "extend", cmd.Use)
}
