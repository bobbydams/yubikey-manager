package ui

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestSetColorEnabled(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor

	// Restore after test
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	t.Run("enable colors", func(t *testing.T) {
		SetColorEnabled(true)
		assert.True(t, IsColorEnabled())
		assert.False(t, color.NoColor)
	})

	t.Run("disable colors", func(t *testing.T) {
		SetColorEnabled(false)
		assert.False(t, IsColorEnabled())
		assert.True(t, color.NoColor)
	})

	t.Run("toggle colors", func(t *testing.T) {
		SetColorEnabled(true)
		assert.True(t, IsColorEnabled())

		SetColorEnabled(false)
		assert.False(t, IsColorEnabled())

		SetColorEnabled(true)
		assert.True(t, IsColorEnabled())
	})
}

func TestIsColorEnabled(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	defer func() {
		colorEnabled = originalEnabled
	}()

	SetColorEnabled(true)
	assert.True(t, IsColorEnabled())

	SetColorEnabled(false)
	assert.False(t, IsColorEnabled())
}

func TestLogInfo(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	t.Run("with colors enabled", func(t *testing.T) {
		SetColorEnabled(true)
		// We can't easily capture stdout in tests, so we just verify the function doesn't panic
		// and that color state is correct
		assert.True(t, IsColorEnabled())
		assert.NotPanics(t, func() {
			LogInfo("test message")
		})
	})

	t.Run("with colors disabled", func(t *testing.T) {
		SetColorEnabled(false)
		assert.False(t, IsColorEnabled())
		assert.NotPanics(t, func() {
			LogInfo("test message")
		})
	})
}

func TestLogSuccess(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		LogSuccess("operation completed")
	})
}

func TestLogWarning(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		LogWarning("this is a warning")
	})
}

func TestLogError(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		LogError("an error occurred")
	})
}

func TestPrintHeader(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	t.Run("with colors", func(t *testing.T) {
		SetColorEnabled(true)
		assert.NotPanics(t, func() {
			PrintHeader("Test Header")
		})
	})

	t.Run("without colors", func(t *testing.T) {
		SetColorEnabled(false)
		assert.NotPanics(t, func() {
			PrintHeader("Test Header")
		})
	})
}

func TestPrintSection(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		PrintSection("SECTION TITLE")
	})
}

func TestPrintKeyValue(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	t.Run("with colors", func(t *testing.T) {
		SetColorEnabled(true)
		assert.NotPanics(t, func() {
			PrintKeyValue("Label", "Value")
		})
	})

	t.Run("without colors", func(t *testing.T) {
		SetColorEnabled(false)
		assert.NotPanics(t, func() {
			PrintKeyValue("Label", "Value")
		})
	})
}

func TestPrintKeyValueKey(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		PrintKeyValueKey("Key ID", "ABC123DEF456")
	})
}

func TestPrintLabel(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		PrintLabel("Label Text")
	})
}

func TestPrintValue(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		PrintValue("Value Text")
	})
}

func TestPrintKey(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	SetColorEnabled(true)
	assert.NotPanics(t, func() {
		PrintKey("ABC123DEF456")
	})
}

func TestColorOutputWithoutANSI(t *testing.T) {
	// Save original state
	originalEnabled := colorEnabled
	originalNoColor := color.NoColor
	defer func() {
		colorEnabled = originalEnabled
		color.NoColor = originalNoColor
	}()

	// Disable colors
	SetColorEnabled(false)
	assert.False(t, IsColorEnabled())
	assert.True(t, color.NoColor)

	// Test all output functions don't panic when colors are disabled
	assert.NotPanics(t, func() {
		LogInfo("test")
		LogSuccess("test")
		PrintHeader("test")
		PrintKeyValue("key", "value")
		PrintKeyValueKey("key", "ABC123")
		PrintLabel("label")
		PrintValue("value")
		PrintKey("ABC123")
		PrintSection("section")
	})
}
