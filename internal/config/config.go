package config

import (
	"os"
	"strconv"
)

// DeploymentMode represents whether the platform runs On-Premise or in Cloud SaaS mode
type DeploymentMode string

const (
	ModeOnPremise DeploymentMode = "on-premise"
	ModeSaaS      DeploymentMode = "saas"
)

// Config holds all configuration values for the Homolog platform
type Config struct {
	DeploymentMode DeploymentMode
	Port           int
	DatabaseURL    string
	StoragePath    string
	StorageType    string // "local" or "s3"
	JWTSecret      string
	SessionExpiryMinutes int
	IdleTimeoutMinutes   int
	MaxUploadSizeBytes   int64

	// SMTP configuration for regulatory alert emails
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Tenant specific for on-premise
	OnPremOrgSlug string
	OnPremOrgName string
}

// Load reads configuration from environment variables with safe enterprise defaults
func Load() *Config {
	mode := DeploymentMode(getEnv("DEPLOYMENT_MODE", string(ModeOnPremise)))
	if mode != ModeSaaS {
		mode = ModeOnPremise
	}

	port, _ := strconv.Atoi(getEnv("PORT", "8090"))
	sessionExpiry, _ := strconv.Atoi(getEnv("SESSION_EXPIRY_MINUTES", "480")) // 8 hours
	idleTimeout, _ := strconv.Atoi(getEnv("IDLE_TIMEOUT_MINUTES", "15"))     // 15 min BPF compliance
	maxUploadMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "1024"), 10, 64) // 1 GB max per file

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		DeploymentMode:       mode,
		Port:                 port,
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5434/homolog?sslmode=disable"),
		StoragePath:          getEnv("STORAGE_PATH", "./data/storage"),
		StorageType:          getEnv("STORAGE_TYPE", "local"),
		JWTSecret:            getEnv("JWT_SECRET", "homolog-default-secret-change-in-production-dz-pharma-2026"),
		SessionExpiryMinutes: sessionExpiry,
		IdleTimeoutMinutes:   idleTimeout,
		MaxUploadSizeBytes:   maxUploadMB * 1024 * 1024,

		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "notifications@homolog.dz"),

		OnPremOrgSlug: getEnv("ONPREM_ORG_SLUG", "laboratoire-principal"),
		OnPremOrgName: getEnv("ONPREM_ORG_NAME", "Laboratoire Pharmaceutique"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
