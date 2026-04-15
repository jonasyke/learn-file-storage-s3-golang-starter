package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"mime"
	"os"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) (string, error) {
	key := make([]byte, 32)
	rand.Read(key)

	encodedName := hex.EncodeToString(key)

	extension, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		return "", fmt.Errorf("could not generate name: %v", err)
	}

	return encodedName + extension[0], nil
}
