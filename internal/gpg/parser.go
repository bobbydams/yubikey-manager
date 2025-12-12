package gpg

import (
	"regexp"
	"strings"
)

// parseKeyList parses the output of `gpg --list-secret-keys`.
func parseKeyList(output []byte) []Key {
	lines := strings.Split(string(output), "\n")
	var keys []Key
	var currentKey *Key

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Primary key: sec   rsa4096/ABC123DEF4567890 2023-01-01 [SC] [expires: 2028-01-01]
		// Subkey:      ssb   ed25519/ABC123... 2023-01-01 [S] [expires: 2028-01-01]
		// Card:         card-no: 0006 12345678
		if strings.HasPrefix(line, "sec") || strings.HasPrefix(line, "ssb") {
			key := parseKeyLine(line)
			keys = append(keys, key)
			currentKey = &keys[len(keys)-1]
		} else if strings.HasPrefix(line, "card-no:") && currentKey != nil {
			// Extract card number
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentKey.CardNo = strings.Join(parts[1:], " ")
			}
		}
	}

	return keys
}

// parseKeyLine parses a single key line from GPG output.
func parseKeyLine(line string) Key {
	key := Key{}

	// Match: sec/ssb   algo/keyid   date   [capabilities] [expires: date]
	re := regexp.MustCompile(`^(sec|ssb)\s+(\S+)/(\S+)\s+(\S+)\s+\[([^\]]+)\](?:\s+\[expires:\s+([^\]]+)\])?`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 6 {
		key.Type = matches[1]
		key.KeyID = matches[3]
		key.Capabilities = parseCapabilities(matches[5])
		if len(matches) >= 7 && matches[6] != "" {
			key.Expires = matches[6]
		}
	}

	return key
}

// parseCapabilities parses capability flags like "[SC]", "[S]", "[E]", "[A]".
func parseCapabilities(caps string) []string {
	var result []string
	for _, char := range caps {
		switch char {
		case 'S':
			result = append(result, "S")
		case 'E':
			result = append(result, "E")
		case 'A':
			result = append(result, "A")
		case 'C':
			result = append(result, "C")
		}
	}
	return result
}

// parseCardStatus parses the output of `gpg --card-status`.
func parseCardStatus(output []byte) *CardInfo {
	info := &CardInfo{
		Keys: make(map[string]string),
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Serial number: 12345678
		if strings.HasPrefix(line, "Serial number") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				info.Serial = parts[3]
			}
		}

		// Name of cardholder: Test User
		if strings.HasPrefix(line, "Name of cardholder") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Cardholder = strings.TrimSpace(parts[1])
			}
		}

		// Signature key....: ABC123... (note the dots for alignment)
		// Match lines like "Signature key.....: ABC123" or "Encryption key....: DEF456"
		if strings.Contains(line, "key") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				// Extract key type (e.g., "Signature key....." -> "Signature")
				keyTypePart := strings.TrimSpace(parts[0])
				// Remove " key" or " key...." suffix (handle variable number of dots)
				keyType := keyTypePart
				// Try to match " key...." pattern (with dots)
				if idx := strings.Index(keyTypePart, " key"); idx > 0 {
					keyType = keyTypePart[:idx]
				}
				keyType = strings.TrimSpace(keyType)
				keyID := strings.TrimSpace(parts[1])
				if keyID != "[none]" && keyID != "" {
					info.Keys[keyType] = keyID
				}
			}
		}
	}

	return info
}
