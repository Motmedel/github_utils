package webhook

import (
	"bytes"
	"errors"
	githubUtilsErrors "github.com/Motmedel/github_utils/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"testing"
)

func TestParseSignature(t *testing.T) {
	testCases := []struct {
		name        string
		signature   string
		expected    []byte
		expectedErr error
	}{
		{
			name:        "empty signature",
			signature:   "",
			expected:    nil,
			expectedErr: githubUtilsErrors.ErrEmptySignature,
		},
		{
			name:        "garbage signature",
			signature:   "garbage",
			expected:    nil,
			expectedErr: githubUtilsErrors.ErrMissingSignatureDelimiter,
		},
		{
			name:        "bad label",
			signature:   "garbage=signature",
			expected:    nil,
			expectedErr: githubUtilsErrors.ErrUnexpectedSignatureLabel,
		},
		{
			name:        "bad signature hex",
			signature:   "sha256=garbage",
			expected:    nil,
			expectedErr: motmedelErrors.ErrSemanticError,
		},
		{
			name:      "valid signature",
			signature: "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17",
			expected: []byte{
				0x75, 0x71, 0x07, 0xea, 0x0e, 0xb2, 0x50, 0x9f,
				0xc2, 0x11, 0x22, 0x1c, 0xce, 0x98, 0x4b, 0x8a,
				0x37, 0x57, 0x0b, 0x6d, 0x75, 0x86, 0xc2, 0x2c,
				0x46, 0xf4, 0x37, 0x9c, 0x8b, 0x04, 0x3e, 0x17,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			signatureBytes, err := ParseSignature(testCase.signature)
			if !bytes.Equal(signatureBytes, testCase.expected) || !errors.Is(err, testCase.expectedErr) {
				t.Fatalf(
					"expected %v and %v, got %v and %v",
					testCase.expected,
					testCase.expectedErr,
					signatureBytes,
					err,
				)
			}
		})
	}
}

func TestValidateSignature(t *testing.T) {
	testCases := []struct {
		name          string
		body          []byte
		webhookSecret []byte
		signature     []byte
		expected      bool
		expectedErr   error
	}{
		{
			name:        "empty body",
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "empty webhook secret",
			body:        []byte("body"),
			expected:    false,
			expectedErr: githubUtilsErrors.ErrEmptyWebhookSecret,
		},
		{
			name:          "empty signature",
			body:          []byte("body"),
			webhookSecret: []byte("secret"),
			expected:      false,
			expectedErr:   githubUtilsErrors.ErrEmptySignature,
		},
		{
			name:          "docs example",
			body:          []byte("Hello, World!"),
			webhookSecret: []byte("It's a Secret to Everybody"),
			signature: []byte{
				0x75, 0x71, 0x07, 0xea, 0x0e, 0xb2, 0x50, 0x9f,
				0xc2, 0x11, 0x22, 0x1c, 0xce, 0x98, 0x4b, 0x8a,
				0x37, 0x57, 0x0b, 0x6d, 0x75, 0x86, 0xc2, 0x2c,
				0x46, 0xf4, 0x37, 0x9c, 0x8b, 0x04, 0x3e, 0x17,
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name:          "unmatching signature",
			body:          []byte("Hello, World!"),
			webhookSecret: []byte("It's a Secret to Everybody"),
			signature: []byte{
				0xff, 0x71, 0x07, 0xea, 0x0e, 0xb2, 0x50, 0x9f,
				0xc2, 0x11, 0x22, 0x1c, 0xce, 0x98, 0x4b, 0x8a,
				0x37, 0x57, 0x0b, 0x6d, 0x75, 0x86, 0xc2, 0x2c,
				0x46, 0xf4, 0x37, 0x9c, 0x8b, 0x04, 0x3e, 0x17,
			},
			expected:    false,
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ok, err := ValidateSignature(testCase.body, testCase.webhookSecret, testCase.signature)
			if ok != testCase.expected || !errors.Is(err, testCase.expectedErr) {
				t.Fatalf(
					"expected %v and %v, got %v and %v",
					testCase.expected,
					testCase.expectedErr,
					ok,
					err,
				)
			}
		})
	}
}
