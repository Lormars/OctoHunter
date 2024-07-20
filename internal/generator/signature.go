package generator

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"

	"github.com/lormars/octohunter/internal/logger"
)

func GenerateSignature() (string, error) {
	b := make([]byte, 7)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		logger.Debugf("Error generating signature: %v\n", err)
		return "", err
	}
	raw := strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
	cleaned := strings.ReplaceAll(raw, "_", "-")
	return cleaned, nil
}
