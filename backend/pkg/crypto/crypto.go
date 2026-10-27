package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

const (
	// KeySize é o tamanho da chave em bytes (32 bytes = 256 bits)
	KeySize = 32
	// NonceSize é o tamanho do nonce em bytes (12 bytes para GCM)
	NonceSize = 12
)

var (
	ErrInvalidKeySize    = errors.New("tamanho de chave inválido: deve ser 32 bytes (256 bits) após decodificação base64")
	ErrEncryptionFailed  = errors.New("falha na criptografia")
	ErrDecryptionFailed  = errors.New("falha na descriptografia")
	ErrInvalidCiphertext = errors.New("texto cifrado inválido")
	ErrKeyNotConfigured  = errors.New("chave de criptografia não configurada: defina INTELLIGATE_ENCRYPTION_KEY")
)

// CryptoService fornece criptografia simétrica AES-256-GCM
type CryptoService struct {
	key []byte
}

// NewCryptoService cria um novo serviço de criptografia
// A chave deve ser uma string base64 que decodifica para 32 bytes
func NewCryptoService(keyB64 string) (*CryptoService, error) {
	if keyB64 == "" {
		return nil, ErrKeyNotConfigured
	}

	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("chave inválida (não é base64 válido): %w", err)
	}

	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}

	return &CryptoService{key: key}, nil
}

// NewCryptoServiceFromEnv cria um CryptoService a partir da variável de ambiente
func NewCryptoServiceFromEnv() (*CryptoService, error) {
	keyB64 := os.Getenv("INTELLIGATE_ENCRYPTION_KEY")
	return NewCryptoService(keyB64)
}

// Encrypt criptografa um texto plano usando AES-256-GCM
// Retorna o texto cifrado em base64 (nonce + ciphertext + tag)
func (c *CryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Combinar nonce + ciphertext para armazenamento
	result := make([]byte, NonceSize+len(ciphertext))
	copy(result[:NonceSize], nonce)
	copy(result[NonceSize:], ciphertext)

	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt descriptografa um texto cifrado em base64
func (c *CryptoService) Decrypt(ciphertextB64 string) (string, error) {
	if ciphertextB64 == "" {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	if len(ciphertext) < NonceSize {
		return "", ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	nonce := ciphertext[:NonceSize]
	ciphertextOnly := ciphertext[NonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertextOnly, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// MustEncrypt criptografa e faz panic em caso de erro (para uso em inicialização)
func (c *CryptoService) MustEncrypt(plaintext string) string {
	result, err := c.Encrypt(plaintext)
	if err != nil {
		panic(fmt.Sprintf("falha crítica na criptografia: %v", err))
	}
	return result
}

// MustDecrypt descriptografa e faz panic em caso de erro (para uso em inicialização)
func (c *CryptoService) MustDecrypt(ciphertextB64 string) string {
	result, err := c.Decrypt(ciphertextB64)
	if err != nil {
		panic(fmt.Sprintf("falha crítica na descriptografia: %v", err))
	}
	return result
}

// GenerateKey gera uma nova chave aleatória de 32 bytes e retorna em base64
// Útil para gerar a chave inicial
func GenerateKey() (string, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
