package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExportCmd(t *testing.T) {
	cmd := newExportCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "export", cmd.Use)
}
