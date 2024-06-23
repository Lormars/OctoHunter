package generator

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/lormars/octohunter/internal/logger"
)

func GenerateSignature() (string, error) {
	b := make([]byte, 7)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		logger.Debugf("Error generating signature: %v\n", err)
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
