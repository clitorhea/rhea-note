package auth

import (
	"os"

	"github.com/zalando/go-keyring"
)

const service = "secnotes"
const user = "master_password"

// GetPassword retrieves the master password from either the environment variable or the OS Keyring.
func GetPassword() (string, error) {
	// 1. Env var takes highest precedence
	if env := os.Getenv("SECNOTES_PASSWORD"); env != "" {
		return env, nil
	}
	// 2. Fall back to secure OS Keyring
	return keyring.Get(service, user)
}

// SavePassword securely stores the master password in the OS Keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager).
func SavePassword(password string) error {
	return keyring.Set(service, user, password)
}

// DeletePassword removes the master password from the OS Keyring.
func DeletePassword() error {
	return keyring.Delete(service, user)
}
