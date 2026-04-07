package config

import (
	"bufio"
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

func Load() Config {
	loadDotEnv(".env")

	return Config{
		AppEnv:        getenv("APP_ENV", "dev"),
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		AppBaseURL:    getenv("APP_BASE_URL", "http://127.0.0.1:8080"),
		SQLitePath:    getenv("SQLITE_PATH", "./data/app.db"),
		AuthSecret:    getenv("AUTH_SECRET", "change-me"),
		AuthLocalUser: getenv("AUTH_LOCAL_USER", "admin"),
		AuthLocalPass: getenv("AUTH_LOCAL_PASS", "admin"),
		LogLevel:      getenv("LOG_LEVEL", "info"),
	}
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
