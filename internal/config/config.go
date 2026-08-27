package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port           string
	AllowedOrigin  string
	AboutMDPath    string
	GitHubToken    string
	GitHubUserName string
	ResendAPIKey   string
	EmailFrom      string
	EmailTo        string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:           getEnvDefault("PORT", "7302"),
		AllowedOrigin:  getEnvDefault("ALLOWED_ORIGIN", "http://localhost:7302"),
		AboutMDPath:    getEnvDefault("ABOUT_MD_PATH", "./about.md"),
		GitHubToken:    os.Getenv("GITHUB_TOKEN"),
		GitHubUserName: os.Getenv("GITHUB_USERNAME"),
		ResendAPIKey:   os.Getenv("RESEND_API_KEY"),
		EmailFrom:      os.Getenv("EMAIL_FROM"),
		EmailTo:        os.Getenv("EMAIL_TO"),
	}

	var missing []string
	if cfg.GitHubToken == "" {
		missing = append(missing, "GITHUB_TOKEN")
	}
	if cfg.GitHubUserName == "" {
		missing = append(missing, "GITHUB_USERNAME")
	}
	if cfg.ResendAPIKey == "" {
		missing = append(missing, "RESEND_API_KEY")
	}
	if cfg.EmailFrom == "" {
		missing = append(missing, "EMAIL_FROM")
	}
	if cfg.EmailTo == "" {
		missing = append(missing, "EMAIL_TO")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %v", missing)
	}

	return cfg, nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
