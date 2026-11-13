package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// DeploymentMode represents whether the platform runs On-Premise or in Cloud SaaS mode
type DeploymentMode string

const (
	ModeOnPremise DeploymentMode = "on-premise"
	ModeSaaS      DeploymentMode = "saas"
)

// Config holds all configuration values for the Homolog platform
type Config struct {
	Environment          string // "development", "staging", "production"
	EnableDevSwitcher    bool   // Guard for development role switcher (disabled in production)
	CookieSecure         bool   // Enforce Secure attribute on session cookies
	DeploymentMode       DeploymentMode
	Port                 int
	DatabaseURL          string
	StoragePath          string
	StorageType          string // "local" or "s3"
	JWTSecret            string
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
	loadDotEnv(".env")

	env := getEnv("APP_ENV", getEnv("ENVIRONMENT", "development"))
	isProd := strings.ToLower(env) == "production" || strings.ToLower(env) == "prod"

	// Dev Switcher is strictly forbidden in production; in non-prod it is true by default unless ENABLE_DEV_SWITCHER=false
	devSwitcherDefault := "true"
	if isProd {
		devSwitcherDefault = "false"
	}
	enableDevSwitch := getEnv("ENABLE_DEV_SWITCHER", devSwitcherDefault) == "true" && !isProd

	// Cookie Secure is true by default in production; can also be forced via COOKIE_SECURE=true
	cookieSecureDefault := "false"
	if isProd {
		cookieSecureDefault = "true"
	}
	cookieSecure := getEnv("COOKIE_SECURE", cookieSecureDefault) == "true"

	mode := DeploymentMode(getEnv("DEPLOYMENT_MODE", string(ModeOnPremise)))
	if mode != ModeSaaS {
		mode = ModeOnPremise
	}

	port, _ := strconv.Atoi(getEnv("PORT", "8090"))
	sessionExpiry, _ := strconv.Atoi(getEnv("SESSION_EXPIRY_MINUTES", "480"))   // 8 hours
	idleTimeout, _ := strconv.Atoi(getEnv("IDLE_TIMEOUT_MINUTES", "15"))        // 15 min BPF compliance
	maxUploadMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "1024"), 10, 64) // 1 GB max per file

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		Environment:          env,
		EnableDevSwitcher:    enableDevSwitch,
		CookieSecure:         cookieSecure,
		DeploymentMode:       mode,
		Port:                 port,
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://homolog:homolog_secret_2026@localhost:5434/homolog?sslmode=disable"),
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
		OnPremOrgName: getEnv("ONPREM_ORG_NAME", "Laboratoire Pharmaceutique National"),
	}
}

// IsProduction returns true if the environment is configured as production
func (c *Config) IsProduction() bool {
	if c == nil {
		return false
	}
	env := strings.ToLower(strings.TrimSpace(c.Environment))
	return env == "production" || env == "prod"
}

// DevSwitcherEnabled returns true only in non-production environments when explicitly enabled
func (c *Config) DevSwitcherEnabled() bool {
	if c == nil {
		appEnv := strings.ToLower(os.Getenv("APP_ENV"))
		if appEnv == "production" || appEnv == "prod" {
			return false
		}
		if os.Getenv("ENABLE_DEV_SWITCHER") == "false" {
			return false
		}
		return true
	}
	return !c.IsProduction() && c.EnableDevSwitcher
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}
