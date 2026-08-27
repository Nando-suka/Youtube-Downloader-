package config

import (
	"crypto/sha256"
	"os"

	"Youtube_donwloader/utils"
)

// getEncryptionKey mendapatkan encryption key dari environment atau generate dari secret
func getEncryptionKey() []byte {
	// Coba ambil dari environment variable
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr != "" {
		// Key harus 32 bytes untuk AES-256
		key := sha256.Sum256([]byte(keyStr))
		return key[:]
	}

	// Fallback: gunakan default key (tidak aman untuk production!)
	// Hanya untuk development/testing
	defaultKey := "youtube-downloader-default-encryption-key-change-in-production"
	key := sha256.Sum256([]byte(defaultKey))
	return key[:]
}

// decryptAPIKeyIfEncrypted mendekripsi API key jika terenkripsi
// Format: jika key dimulai dengan "encrypted:", maka didekripsi
func decryptAPIKeyIfEncrypted(encryptedKey string) (string, error) {
	if len(encryptedKey) == 0 {
		return "", nil
	}

	// Cek apakah key terenkripsi (format: "encrypted:base64string")
	const prefix = "encrypted:"
	if len(encryptedKey) > len(prefix) && encryptedKey[:len(prefix)] == prefix {
		encrypted := encryptedKey[len(prefix):]
		key := getEncryptionKey()
		decrypted, err := utils.DecryptAPIKey(encrypted, key)
		if err != nil {
			return "", err
		}
		return decrypted, nil
	}

	// Jika tidak terenkripsi, return as-is
	return encryptedKey, nil
}

// EncryptAPIKeyForStorage mengenkripsi API key untuk disimpan
// Ini helper function untuk encrypt key sebelum disimpan ke file/env
func EncryptAPIKeyForStorage(apiKey string) (string, error) {
	key := getEncryptionKey()
	encrypted, err := utils.EncryptAPIKey(apiKey, key)
	if err != nil {
		return "", err
	}
	return "encrypted:" + encrypted, nil
}
