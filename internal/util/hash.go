package util

import (
	"crypto/sha256"
	"fmt"
)

// GenerateHash creates a SHA256 hash from the given content and image data.
func GenerateHash(content string, imageData []byte) string {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	if imageData != nil {
		hasher.Write(imageData)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
