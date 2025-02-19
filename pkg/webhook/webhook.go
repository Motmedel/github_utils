package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	githubUtilsErrors "github.com/Motmedel/github_utils/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"strings"
)

const SignatureHeaderName = "X-Hub-Signature-256"

func ParseSignature(signature string) ([]byte, error) {
	if signature == "" {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrSyntaxError, githubUtilsErrors.ErrEmptySignature),
		)
	}

	label, signatureHex, found := strings.Cut(signature, "=")
	if !found {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrSyntaxError, githubUtilsErrors.ErrMissingSignatureDelimiter),
			signature,
		)
	}

	if label != "sha256" {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf(
				"%w: %w: %s",
				motmedelErrors.ErrSemanticError,
				githubUtilsErrors.ErrUnexpectedSignatureLabel,
				label,
			),
			label,
		)
	}

	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("%w: hex decode string: %w", motmedelErrors.ErrSemanticError, err),
			signatureHex,
		)
	}

	// TODO: Check bytes length?

	return signatureBytes, nil
}

func ValidateSignature(body []byte, webhookSecret []byte, signature []byte) (bool, error) {
	if len(body) == 0 {
		return false, nil
	}

	if len(webhookSecret) == 0 {
		return false, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptyWebhookSecret)
	}

	if len(signature) == 0 {
		return false, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptySignature)
	}

	mac := hmac.New(sha256.New, webhookSecret)
	mac.Write(body)
	ourSignature := mac.Sum(nil)

	return hmac.Equal(ourSignature, signature), nil
}
