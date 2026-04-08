package config

import "testing"

func TestValidateRejectsInsecureAuthDefaults(t *testing.T) {
	cfg := Config{
		AuthSecret:    "change-me",
		AuthLocalUser: "admin",
		AuthLocalPass: "admin",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for insecure auth defaults")
	}
}

func TestValidateAcceptsExplicitAuthConfig(t *testing.T) {
	cfg := Config{
		AuthSecret:    "test-secret",
		AuthLocalUser: "admin",
		AuthLocalPass: "dev-password-123",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}
