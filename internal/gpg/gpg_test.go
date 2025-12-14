package gpg

import (
	"context"
	"testing"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ListSecretKeys(t *testing.T) {
	tests := []struct {
		name          string
		keyID         string
		mockOutput    string
		mockError     error
		expectedKeys  int
		expectedError bool
	}{
		{
			name:  "successful list",
			keyID: "ABC123DEF4567890",
			mockOutput: `sec   rsa4096/ABC123DEF4567890 2023-01-01 [SC] [expires: 2028-01-01]
uid                 [ultimate] Test User <test@example.com>
ssb   ed25519/ABC123DEF456 2023-01-01 [S] [expires: 2028-01-01]
card-no: 0006 12345678
`,
			expectedKeys:  2,
			expectedError: false,
		},
		{
			name:          "key not found",
			keyID:         "INVALID",
			mockOutput:    "",
			mockError:     nil,
			expectedKeys:  0,
			expectedError: false,
		},
		{
			name:  "keys on card (sec# and ssb>)",
			keyID: "07AAA1E535650AF5",
			mockOutput: `sec#  ed25519/07AAA1E535650AF5 2025-09-05 [SC] [expires: 2030-09-04]
      FA57C85131F11B28EE236A4F07AAA1E535650AF5
uid                 [ultimate] Test User <test@example.com>
ssb>  cv25519/116DB85718F8B287 2025-09-05 [E] [expires: 2030-09-04]
ssb>  ed25519/DC47D1B090A51498 2025-09-05 [S] [expires: 2030-09-04]
ssb>  ed25519/0257F6B8152D7F35 2025-09-05 [A] [expires: 2030-09-04]
`,
			expectedKeys:  4,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := executor.NewMockExecutor()
			svc := NewService(mockExec)

			key := "gpg --list-secret-keys --keyid-format=long " + tt.keyID
			mockExec.SetOutput(key, []byte(tt.mockOutput))
			if tt.mockError != nil {
				mockExec.SetError(key, tt.mockError)
			}

			keys, err := svc.ListSecretKeys(context.Background(), tt.keyID)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, keys, tt.expectedKeys)
			}
		})
	}
}

func TestService_CardStatus(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		expectedSerial string
		expectedError  bool
	}{
		{
			name: "successful card status",
			mockOutput: `Reader ...........: Yubico YubiKey OTP FIDO CCID
Application ID ...: D2760001240102010006055532110000
Version ..........: 5.4.3
Serial number ....: 12345678
			Name of cardholder: Test User
Signature key ....: ABC123DEF4567890
Encryption key....: DEF456GHI7890123
Authentication key: GHI789JKL0123456
`,
			expectedSerial: "12345678",
			expectedError:  false,
		},
		{
			name:          "no card present",
			mockOutput:    "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := executor.NewMockExecutor()
			svc := NewService(mockExec)

			key := "gpg --card-status"
			mockExec.SetOutput(key, []byte(tt.mockOutput))
			if tt.expectedError {
				mockExec.SetError(key, assert.AnError)
			}

			cardInfo, err := svc.CardStatus(context.Background())

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSerial, cardInfo.Serial)
			}
		})
	}
}

func TestService_ExportPublicKey(t *testing.T) {
	mockExec := executor.NewMockExecutor()
	svc := NewService(mockExec)

	keyID := "ABC123DEF4567890"
	expectedOutput := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n...\n-----END PGP PUBLIC KEY BLOCK-----")

	key := "gpg --export --armor " + keyID
	mockExec.SetOutput(key, expectedOutput)

	output, err := svc.ExportPublicKey(context.Background(), keyID)

	require.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}

func TestParseKeyList(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		checkCardNo bool
	}{
		{
			name: "parse keys with card",
			input: `sec   rsa4096/ABC123DEF4567890 2023-01-01 [SC] [expires: 2028-01-01]
uid                 [ultimate] Test User <test@example.com>
ssb   ed25519/ABC123DEF456 2023-01-01 [S] [expires: 2028-01-01]
card-no: 0006 12345678
`,
			expectedLen: 2,
			checkCardNo: true,
		},
		{
			name: "parse keys without card",
			input: `sec   rsa4096/ABC123DEF4567890 2023-01-01 [SC] [expires: 2028-01-01]
ssb   ed25519/ABC123DEF456 2023-01-01 [S] [expires: 2028-01-01]
`,
			expectedLen: 2,
			checkCardNo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := parseKeyList([]byte(tt.input))
			assert.Len(t, keys, tt.expectedLen)

			if tt.checkCardNo {
				foundCardNo := false
				for _, key := range keys {
					if key.CardNo != "" {
						foundCardNo = true
						break
					}
				}
				assert.True(t, foundCardNo, "expected at least one key with card-no")
			}
		})
	}
}

func TestParseCardStatus(t *testing.T) {
	input := `Reader ...........: Yubico YubiKey OTP FIDO CCID
Application ID ...: D2760001240102010006055532110000
Version ..........: 5.4.3
Serial number ....: 12345678
			Name of cardholder: Test User
Signature key.....: ABC123DEF4567890
Encryption key....: DEF456GHI7890123
Authentication key: GHI789JKL0123456
`

	cardInfo := parseCardStatus([]byte(input))

	assert.Equal(t, "12345678", cardInfo.Serial)
	assert.Equal(t, "Test User", cardInfo.Cardholder)
	assert.Equal(t, "ABC123DEF4567890", cardInfo.Keys["Signature"])
	assert.Equal(t, "DEF456GHI7890123", cardInfo.Keys["Encryption"])
	assert.Equal(t, "GHI789JKL0123456", cardInfo.Keys["Authentication"])
}
