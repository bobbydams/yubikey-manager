package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetadataCmd(t *testing.T) {
	cmd := newMetadataCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "set-metadata", cmd.Use)
}
