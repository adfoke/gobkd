package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv        string
	HTTPAddr      string
	AppBaseURL    string
	SQLitePath    string
	AuthSecret    string
	AuthLocalUser string
	AuthLocalPass string
	LogLevel      string
}

func Load() (Config, error) {
	loadDotEnv(".env")

	cfg := Config{
		AppEnv:        getenv("APP_ENV", "dev"),
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		AppBaseURL:    getenv("APP_BASE_URL", "http://127.0.0.1:8080"),
		SQLitePath:    getenv("SQLITE_PATH", "./data/app.db"),
		AuthSecret:    getenv("AUTH_SECRET", ""),
		AuthLocalUser: getenv("AUTH_LOCAL_USER", ""),
		AuthLocalPass: getenv("AUTH_LOCAL_PASS", ""),
		LogLevel:      getenv("LOG_LEVEL", "info"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	var problems []string

	switch {
	case strings.TrimSpace(c.AuthSecret) == "":
		problems = append(problems, "AUTH_SECRET is required")
	case c.AuthSecret == "change-me":
		problems = append(problems, "AUTH_SECRET must not use the placeholder value")
	}

	if strings.TrimSpace(c.AuthLocalUser) == "" {
		problems = append(problems, "AUTH_LOCAL_USER is required")
	}

	switch {
	case strings.TrimSpace(c.AuthLocalPass) == "":
		problems = append(problems, "AUTH_LOCAL_PASS is required")
	case c.AuthLocalPass == "admin":
		problems = append(problems, "AUTH_LOCAL_PASS must not use the placeholder value")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid auth config: %s", strings.Join(problems, "; "))
	}

	return nil
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
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

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		_ = os.Setenv(key, value)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
