package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user for a yes/no confirmation.
// Returns true if the user responds with 'y' or 'yes' (case-insensitive).
// Returns false for any other response or empty input.
func Confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N] ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// Prompt reads a line of input from the user.
// Returns the trimmed input string.
func Prompt(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// PromptRequired reads a line of input from the user and ensures it's not empty.
// Continues prompting until a non-empty response is provided.
func PromptRequired(prompt string) (string, error) {
	for {
		response, err := Prompt(prompt)
		if err != nil {
			return "", err
		}

		if response != "" {
			return response, nil
		}

		LogWarning("This field is required. Please enter a value.")
	}
}
