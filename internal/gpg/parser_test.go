package gpg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "sign only",
			input:    "S",
			expected: []string{"S"},
		},
		{
			name:     "sign and encrypt",
			input:    "SE",
			expected: []string{"S", "E"},
		},
		{
			name:     "sign, certify, encrypt",
			input:    "SCE",
			expected: []string{"S", "C", "E"},
		},
		{
			name:     "all capabilities",
			input:    "SCEA",
			expected: []string{"S", "C", "E", "A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCapabilities(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseKeyLine(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  string
		expectedKeyID string
		hasExpires    bool
	}{
		{
			name:          "primary key with expiration",
			input:         "sec   rsa4096/ABC123DEF4567890 2023-01-01 [SC] [expires: 2028-01-01]",
			expectedType:  "sec",
			expectedKeyID: "ABC123DEF4567890",
			hasExpires:    true,
		},
		{
			name:          "subkey without expiration",
			input:         "ssb   ed25519/ABC123DEF456 2023-01-01 [S]",
			expectedType:  "ssb",
			expectedKeyID: "ABC123DEF456",
			hasExpires:    false,
		},
		{
			name:          "primary key on card (sec#)",
			input:         "sec#  ed25519/07AAA1E535650AF5 2025-09-05 [SC] [expires: 2030-09-04]",
			expectedType:  "sec",
			expectedKeyID: "07AAA1E535650AF5",
			hasExpires:    true,
		},
		{
			name:          "subkey on card (ssb>)",
			input:         "ssb>  ed25519/DC47D1B090A51498 2025-09-05 [S] [expires: 2030-09-04]",
			expectedType:  "ssb",
			expectedKeyID: "DC47D1B090A51498",
			hasExpires:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := parseKeyLine(tt.input)
			assert.Equal(t, tt.expectedType, key.Type)
			assert.Equal(t, tt.expectedKeyID, key.KeyID)
			if tt.hasExpires {
				assert.NotEmpty(t, key.Expires)
			}
		})
	}
}
