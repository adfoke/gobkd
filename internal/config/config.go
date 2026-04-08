package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                string
	HTTPAddr              string
	AppBaseURL            string
	SQLitePath            string
	AuthSecret            string
	AuthLocalUser         string
	AuthLocalPass         string
	LogLevel              string
	HTTPReadTimeout       time.Duration
	HTTPReadHeaderTimeout time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	HTTPShutdownTimeout   time.Duration
	HTTPMaxHeaderBytes    int
	HTTPMaxBodyBytes      int64
	HTTPTrustedProxies    []string
}

func Load() (Config, error) {
	loadDotEnv(".env")

	readTimeout, err := getDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	idleTimeout, err := getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := getDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	maxHeaderBytes, err := getInt("HTTP_MAX_HEADER_BYTES", 1<<20)
	if err != nil {
		return Config{}, err
	}

	maxBodyBytes, err := getInt64("HTTP_MAX_BODY_BYTES", 1<<20)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:                getenv("APP_ENV", "dev"),
		HTTPAddr:              getenv("HTTP_ADDR", ":8080"),
		AppBaseURL:            getenv("APP_BASE_URL", "http://127.0.0.1:8080"),
		SQLitePath:            getenv("SQLITE_PATH", "./data/app.db"),
		AuthSecret:            getenv("AUTH_SECRET", ""),
		AuthLocalUser:         getenv("AUTH_LOCAL_USER", ""),
		AuthLocalPass:         getenv("AUTH_LOCAL_PASS", ""),
		LogLevel:              getenv("LOG_LEVEL", "info"),
		HTTPReadTimeout:       readTimeout,
		HTTPReadHeaderTimeout: readHeaderTimeout,
		HTTPWriteTimeout:      writeTimeout,
		HTTPIdleTimeout:       idleTimeout,
		HTTPShutdownTimeout:   shutdownTimeout,
		HTTPMaxHeaderBytes:    maxHeaderBytes,
		HTTPMaxBodyBytes:      maxBodyBytes,
		HTTPTrustedProxies:    getCSV("HTTP_TRUSTED_PROXIES"),
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

	if c.HTTPReadTimeout <= 0 {
		problems = append(problems, "HTTP_READ_TIMEOUT must be greater than 0")
	}
	if c.HTTPReadHeaderTimeout <= 0 {
		problems = append(problems, "HTTP_READ_HEADER_TIMEOUT must be greater than 0")
	}
	if c.HTTPWriteTimeout <= 0 {
		problems = append(problems, "HTTP_WRITE_TIMEOUT must be greater than 0")
	}
	if c.HTTPIdleTimeout <= 0 {
		problems = append(problems, "HTTP_IDLE_TIMEOUT must be greater than 0")
	}
	if c.HTTPShutdownTimeout <= 0 {
		problems = append(problems, "HTTP_SHUTDOWN_TIMEOUT must be greater than 0")
	}
	if c.HTTPMaxHeaderBytes <= 0 {
		problems = append(problems, "HTTP_MAX_HEADER_BYTES must be greater than 0")
	}
	if c.HTTPMaxBodyBytes <= 0 {
		problems = append(problems, "HTTP_MAX_BODY_BYTES must be greater than 0")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(problems, "; "))
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

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return value, nil
}

func getInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	return value, nil
}

func getInt64(key string, fallback int64) (int64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	return value, nil
}

func getCSV(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return []string{}
	}

	items := strings.Split(raw, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}

	return out
}
