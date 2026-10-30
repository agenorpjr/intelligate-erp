package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	AppEnv             string
	AppPort            string
	AppVersion         string
	HTTPTimeout        time.Duration
	JWTSecret          string
	JWTExpiry          time.Duration
	EncryptionKey      string
	EncryptionKeyBytes []byte

	// Configuração do provedor de IA (Etapa 13).
	// AIAPIKey nunca deve ser registrada em logs.
	AIProvider string
	AIAPIKey   string
	AIModel    string
	AIBaseURL  string

	// Exibir token de redefinição de senha em desenvolvimento (Etapa 16).
	// Em produção deve ser false; o token é enviado por e-mail.
	ExibirTokenRedefinicaoDev bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	timeoutStr := getEnv("HTTP_TIMEOUT", "30s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 30 * time.Second
	}

	jwtExpiryStr := getEnv("JWT_EXPIRY", "24h")
	jwtExpiry, err := time.ParseDuration(jwtExpiryStr)
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	encryptionKey := getEnv("INTELLIGATE_ENCRYPTION_KEY", "")
	var encryptionKeyBytes []byte
	if encryptionKey != "" {
		// Validar que a chave é base64 válido e tem 32 bytes
		decoded, err := base64.StdEncoding.DecodeString(encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("INTELLIGATE_ENCRYPTION_KEY inválido (não é base64): %w", err)
		}
		if len(decoded) != 32 {
			return nil, fmt.Errorf("INTELLIGATE_ENCRYPTION_KEY inválido: deve decodificar para 32 bytes (256 bits), got %d", len(decoded))
		}
		encryptionKeyBytes = decoded
	}

	return &Config{
		DatabaseURL:               getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/intelligate?sslmode=disable"),
		AppEnv:                    getEnv("APP_ENV", "development"),
		AppPort:                   getEnv("APP_PORT", "8080"),
		AppVersion:                getEnv("APP_VERSION", "0.1.0"),
		HTTPTimeout:               timeout,
		JWTSecret:                 getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:                 jwtExpiry,
		EncryptionKey:             encryptionKey,
		EncryptionKeyBytes:        encryptionKeyBytes,
		AIProvider:                getEnv("INTELLIGATE_AI_PROVIDER", "mock"),
		AIAPIKey:                  getEnv("INTELLIGATE_AI_API_KEY", ""),
		AIModel:                   getEnv("INTELLIGATE_AI_MODEL", ""),
		AIBaseURL:                 getEnv("INTELLIGATE_AI_BASE_URL", ""),
		ExibirTokenRedefinicaoDev: GetEnvAsBool("EXIBIR_TOKEN_REDEFINICAO_DEV", false),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
