package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	// colorEnabled controls whether colors are used
	colorEnabled = true

	// InfoColor is used for informational messages
	InfoColor = color.New(color.FgBlue)
	// SuccessColor is used for success messages
	SuccessColor = color.New(color.FgGreen)
	// WarningColor is used for warning messages
	WarningColor = color.New(color.FgYellow)
	// ErrorColor is used for error messages
	ErrorColor = color.New(color.FgRed)
	// HeaderColor is used for section headers
	HeaderColor = color.New(color.FgCyan, color.Bold)
	// LabelColor is used for labels
	LabelColor = color.New(color.FgWhite)
	// ValueColor is used for values
	ValueColor = color.New(color.FgHiWhite)
	// KeyColor is used for key IDs and fingerprints
	KeyColor = color.New(color.FgMagenta)
)

// SetColorEnabled enables or disables color output globally
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
	if !enabled {
		// Disable all colors
		color.NoColor = true
	} else {
		// Re-enable colors
		color.NoColor = false
	}
}

// IsColorEnabled returns whether colors are currently enabled
func IsColorEnabled() bool {
	return colorEnabled
}

// LogInfo prints an informational message with [INFO] prefix.
func LogInfo(format string, args ...interface{}) {
	InfoColor.Fprintf(os.Stdout, "[INFO] %s\n", fmt.Sprintf(format, args...))
}

// LogSuccess prints a success message with [SUCCESS] prefix.
func LogSuccess(format string, args ...interface{}) {
	SuccessColor.Fprintf(os.Stdout, "[SUCCESS] %s\n", fmt.Sprintf(format, args...))
}

// LogWarning prints a warning message with [WARNING] prefix.
func LogWarning(format string, args ...interface{}) {
	WarningColor.Fprintf(os.Stderr, "[WARNING] %s\n", fmt.Sprintf(format, args...))
}

// LogError prints an error message with [ERROR] prefix.
func LogError(format string, args ...interface{}) {
	ErrorColor.Fprintf(os.Stderr, "[ERROR] %s\n", fmt.Sprintf(format, args...))
}

// PrintHeader prints a formatted header section with color.
func PrintHeader(title string) {
	fmt.Println()
	HeaderColor.Println("========================================")
	HeaderColor.Printf("       %s\n", title)
	HeaderColor.Println("========================================")
	fmt.Println()
}

// PrintLabel prints a label with color.
func PrintLabel(label string) {
	LabelColor.Print(label)
}

// PrintValue prints a value with color.
func PrintValue(value string) {
	ValueColor.Print(value)
}

// PrintKey prints a key ID or fingerprint with color.
func PrintKey(key string) {
	KeyColor.Print(key)
}

// PrintSection prints a section title with color.
func PrintSection(title string) {
	fmt.Println()
	HeaderColor.Printf("%s\n", title)
	HeaderColor.Println(strings.Repeat("-", len(title)))
	fmt.Println()
}

// PrintKeyValue prints a key-value pair with colors.
func PrintKeyValue(key, value string) {
	LabelColor.Printf("%-25s ", key+":")
	ValueColor.Println(value)
}

// PrintKeyValueKey prints a key-value pair where the value is a key ID/fingerprint.
func PrintKeyValueKey(key, value string) {
	LabelColor.Printf("%-25s ", key+":")
	KeyColor.Println(value)
}
