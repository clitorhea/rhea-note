package crypto

import (
	"bytes"
	"testing"
)

func TestCrypto(t *testing.T) {
	password := "supersecret"
	
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}
	
	key := DeriveKey(password, salt)
	if len(key) != KeySize {
		t.Fatalf("expected key length %d, got %d", KeySize, len(key))
	}
	
	plaintext := []byte("this is a secure note")
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}
	
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}
	
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("expected %s, got %s", plaintext, decrypted)
	}
	
	// Test tampering
	ciphertext[len(ciphertext)-1] ^= 1
	_, err = Decrypt(ciphertext, key)
	if err == nil {
		t.Fatal("expected error when decrypting tampered ciphertext")
	}
}
