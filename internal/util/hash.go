package util

import (
	"crypto/sha256"
	"fmt"
)

func GenerateHash(content string, imageData []byte) string {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	if imageData != nil {
		hasher.Write(imageData)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
