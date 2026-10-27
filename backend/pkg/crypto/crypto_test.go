package crypto

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if key == "" {
		t.Fatal("GenerateKey() returned empty key")
	}

	// Verificar que é base64 válido e tem 32 bytes
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Fatalf("Generated key is not valid base64: %v", err)
	}

	if len(decoded) != KeySize {
		t.Fatalf("Generated key has wrong size: got %d, want %d", len(decoded), KeySize)
	}
}

func TestNewCryptoService(t *testing.T) {
	key, _ := GenerateKey()

	// Teste com chave válida
	cs, err := NewCryptoService(key)
	if err != nil {
		t.Fatalf("NewCryptoService() error = %v", err)
	}
	if cs == nil {
		t.Fatal("NewCryptoService() returned nil")
	}

	// Teste com chave vazia
	_, err = NewCryptoService("")
	if err == nil {
		t.Fatal("NewCryptoService() should fail with empty key")
	}
	if err != ErrKeyNotConfigured {
		t.Fatalf("NewCryptoService() wrong error: got %v, want %v", err, ErrKeyNotConfigured)
	}

	// Teste com chave inválida (não base64)
	_, err = NewCryptoService("not-base64!")
	if err == nil {
		t.Fatal("NewCryptoService() should fail with invalid base64")
	}

	// Teste com chave de tamanho errado
	shortKey := base64.StdEncoding.EncodeToString([]byte("short"))
	_, err = NewCryptoService(shortKey)
	if err == nil {
		t.Fatal("NewCryptoService() should fail with short key")
	}
	if err != ErrInvalidKeySize {
		t.Fatalf("NewCryptoService() wrong error for short key: got %v, want %v", err, ErrInvalidKeySize)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, _ := GenerateKey()
	cs, _ := NewCryptoService(key)

	testCases := []string{
		"",
		"simple",
		"sk_test_1234567890abcdef",
		"client_secret_with_special_chars!@#$%^&*()",
		"very_long_secret_" + string(make([]byte, 1000)),
		"unicode_测试_🔐_секрет",
	}

	for _, tc := range testCases {
		t.Run(tc[:min(20, len(tc))], func(t *testing.T) {
			encrypted, err := cs.Encrypt(tc)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if tc == "" {
				if encrypted != "" {
					t.Fatalf("Encrypt('') should return empty string, got %q", encrypted)
				}
				return
			}

			if encrypted == tc {
				t.Fatal("Encrypt() returned plaintext")
			}

			decrypted, err := cs.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tc {
				t.Fatalf("Decrypt(Encrypt(x)) != x: got %q, want %q", decrypted, tc)
			}
		})
	}
}

func TestEncryptDecryptDifferentKeys(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	cs1, _ := NewCryptoService(key1)
	cs2, _ := NewCryptoService(key2)

	plaintext := "test_secret_123"
	encrypted, _ := cs1.Encrypt(plaintext)

	// Descriptografar com chave diferente deve falhar
	_, err := cs2.Decrypt(encrypted)
	if err == nil {
		t.Fatal("Decrypt with different key should fail")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key, _ := GenerateKey()
	cs, _ := NewCryptoService(key)

	// Ciphertext muito curto
	_, err := cs.Decrypt("dG9vX3Nob3J0") // "too_short" em base64
	if err == nil {
		t.Fatal("Decrypt() should fail with short ciphertext")
	}
	if err != ErrInvalidCiphertext {
		t.Fatalf("Wrong error: got %v, want %v", err, ErrInvalidCiphertext)
	}

	// Ciphertext inválido (base64 válido mas não criptografado corretamente)
	_, err = cs.Decrypt("dGhpcyBpcyBub3QgYSB2YWxpZCBjaXBoZXJ0ZXh0")
	if err == nil {
		t.Fatal("Decrypt() should fail with invalid ciphertext")
	}
}

func TestMustEncryptMustDecrypt(t *testing.T) {
	key, _ := GenerateKey()
	cs, _ := NewCryptoService(key)

	plaintext := "test_secret"
	encrypted := cs.MustEncrypt(plaintext)
	decrypted := cs.MustDecrypt(encrypted)

	if decrypted != plaintext {
		t.Fatalf("MustDecrypt(MustEncrypt(x)) != x: got %q, want %q", decrypted, plaintext)
	}
}

func TestNewCryptoServiceFromEnv(t *testing.T) {
	key, _ := GenerateKey()

	// Definir variável de ambiente
	os.Setenv("INTELLIGATE_ENCRYPTION_KEY", key)
	defer os.Unsetenv("INTELLIGATE_ENCRYPTION_KEY")

	cs, err := NewCryptoServiceFromEnv()
	if err != nil {
		t.Fatalf("NewCryptoServiceFromEnv() error = %v", err)
	}
	if cs == nil {
		t.Fatal("NewCryptoServiceFromEnv() returned nil")
	}

	// Testar sem variável de ambiente
	os.Unsetenv("INTELLIGATE_ENCRYPTION_KEY")
	_, err = NewCryptoServiceFromEnv()
	if err == nil {
		t.Fatal("NewCryptoServiceFromEnv() should fail without env var")
	}
	if err != ErrKeyNotConfigured {
		t.Fatalf("Wrong error: got %v, want %v", err, ErrKeyNotConfigured)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
