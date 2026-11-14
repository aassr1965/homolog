package config

import (
	"os"
	"testing"
)

func TestConfig_ProductionGuards(t *testing.T) {
	// Test case 1: Default in development
	os.Clearenv()
	cfg := Load()
	if cfg.IsProduction() {
		t.Errorf("Expected IsProduction() to be false by default, got true")
	}
	if !cfg.DevSwitcherEnabled() {
		t.Errorf("Expected DevSwitcherEnabled() to be true in dev by default, got false")
	}
	if cfg.CookieSecure {
		t.Errorf("Expected CookieSecure to be false in dev by default, got true")
	}

	// Test case 2: Production environment explicitly sets guards
	os.Setenv("APP_ENV", "production")
	os.Setenv("ENABLE_DEV_SWITCHER", "true") // Attempt to bypass in prod
	cfgProd := Load()
	if !cfgProd.IsProduction() {
		t.Errorf("Expected IsProduction() to be true")
	}
	if cfgProd.DevSwitcherEnabled() {
		t.Errorf("Expected DevSwitcherEnabled() to be strictly false in production even if ENABLE_DEV_SWITCHER=true")
	}
	if !cfgProd.CookieSecure {
		t.Errorf("Expected CookieSecure to be true in production by default")
	}

	// Test case 3: Development with COOKIE_SECURE override
	os.Setenv("APP_ENV", "development")
	os.Setenv("COOKIE_SECURE", "true")
	cfgDevSecure := Load()
	if !cfgDevSecure.CookieSecure {
		t.Errorf("Expected CookieSecure to be true when forced by env")
	}

	// Test case 4: Explicitly disabled in staging/dev
	os.Setenv("APP_ENV", "staging")
	os.Setenv("ENABLE_DEV_SWITCHER", "false")
	cfgStaging := Load()
	if cfgStaging.DevSwitcherEnabled() {
		t.Errorf("Expected DevSwitcherEnabled() to be false when set to false")
	}

	// Cleanup
	os.Clearenv()
}
