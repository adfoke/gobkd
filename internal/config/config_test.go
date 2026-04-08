package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidateRejectsInsecureAuthDefaults(t *testing.T) {
	cfg := Config{
		AuthSecret:            "change-me",
		AuthLocalUser:         "admin",
		AuthLocalPass:         "admin",
		HTTPReadTimeout:       time.Second,
		HTTPReadHeaderTimeout: time.Second,
		HTTPWriteTimeout:      time.Second,
		HTTPIdleTimeout:       time.Second,
		HTTPShutdownTimeout:   time.Second,
		HTTPMaxHeaderBytes:    1024,
		HTTPMaxBodyBytes:      1024,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for insecure auth defaults")
	}
}

func TestValidateAcceptsExplicitAuthConfig(t *testing.T) {
	cfg := Config{
		AuthSecret:            "test-secret",
		AuthLocalUser:         "admin",
		AuthLocalPass:         "dev-password-123",
		HTTPReadTimeout:       time.Second,
		HTTPReadHeaderTimeout: time.Second,
		HTTPWriteTimeout:      time.Second,
		HTTPIdleTimeout:       time.Second,
		HTTPShutdownTimeout:   time.Second,
		HTTPMaxHeaderBytes:    1024,
		HTTPMaxBodyBytes:      1024,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateRejectsInvalidHTTPRuntimeConfig(t *testing.T) {
	cfg := Config{
		AuthSecret:            "test-secret",
		AuthLocalUser:         "admin",
		AuthLocalPass:         "dev-password-123",
		HTTPReadHeaderTimeout: time.Second,
		HTTPWriteTimeout:      time.Second,
		HTTPIdleTimeout:       time.Second,
		HTTPShutdownTimeout:   time.Second,
		HTTPMaxHeaderBytes:    1024,
		HTTPMaxBodyBytes:      1024,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid runtime config")
	}
	if !strings.Contains(err.Error(), "HTTP_READ_TIMEOUT") {
		t.Fatalf("expected runtime config error, got %v", err)
	}
}

func TestLoadParsesHTTPRuntimeConfig(t *testing.T) {
	t.Setenv("AUTH_SECRET", "test-secret")
	t.Setenv("AUTH_LOCAL_USER", "admin")
	t.Setenv("AUTH_LOCAL_PASS", "dev-password-123")
	t.Setenv("HTTP_READ_TIMEOUT", "11s")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "3s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "17s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "90s")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "12s")
	t.Setenv("HTTP_MAX_HEADER_BYTES", "2048")
	t.Setenv("HTTP_MAX_BODY_BYTES", "4096")
	t.Setenv("HTTP_TRUSTED_PROXIES", "127.0.0.1,10.0.0.0/8")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected load success, got %v", err)
	}

	if cfg.HTTPReadTimeout != 11*time.Second {
		t.Fatalf("expected read timeout 11s, got %s", cfg.HTTPReadTimeout)
	}
	if cfg.HTTPReadHeaderTimeout != 3*time.Second {
		t.Fatalf("expected read header timeout 3s, got %s", cfg.HTTPReadHeaderTimeout)
	}
	if cfg.HTTPWriteTimeout != 17*time.Second {
		t.Fatalf("expected write timeout 17s, got %s", cfg.HTTPWriteTimeout)
	}
	if cfg.HTTPIdleTimeout != 90*time.Second {
		t.Fatalf("expected idle timeout 90s, got %s", cfg.HTTPIdleTimeout)
	}
	if cfg.HTTPShutdownTimeout != 12*time.Second {
		t.Fatalf("expected shutdown timeout 12s, got %s", cfg.HTTPShutdownTimeout)
	}
	if cfg.HTTPMaxHeaderBytes != 2048 {
		t.Fatalf("expected max header bytes 2048, got %d", cfg.HTTPMaxHeaderBytes)
	}
	if cfg.HTTPMaxBodyBytes != 4096 {
		t.Fatalf("expected max body bytes 4096, got %d", cfg.HTTPMaxBodyBytes)
	}
	if len(cfg.HTTPTrustedProxies) != 2 {
		t.Fatalf("expected 2 trusted proxies, got %d", len(cfg.HTTPTrustedProxies))
	}
}
