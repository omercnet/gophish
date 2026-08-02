package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func applyEnvironment(config *Config) error {
	setString("ADMIN_LISTEN_URL", &config.AdminConf.ListenURL)
	if err := setBool("ADMIN_USE_TLS", &config.AdminConf.UseTLS); err != nil {
		return err
	}
	setString("ADMIN_CERT_PATH", &config.AdminConf.CertPath)
	setString("ADMIN_KEY_PATH", &config.AdminConf.KeyPath)
	setString("ADMIN_CSRF_KEY", &config.AdminConf.CSRFKey)
	setCSV("ADMIN_ALLOWED_INTERNAL_HOSTS", &config.AdminConf.AllowedInternalHosts)
	setCSV("ADMIN_TRUSTED_ORIGINS", &config.AdminConf.TrustedOrigins)

	setString("PHISH_LISTEN_URL", &config.PhishConf.ListenURL)
	if err := setBool("PHISH_USE_TLS", &config.PhishConf.UseTLS); err != nil {
		return err
	}
	setString("PHISH_CERT_PATH", &config.PhishConf.CertPath)
	setString("PHISH_KEY_PATH", &config.PhishConf.KeyPath)

	setString("DB_NAME", &config.DBName)
	setString("DB_FILE_PATH", &config.DBPath)
	setString("DB_PATH", &config.DBPath)
	setString("DB_SSLCA_PATH", &config.DBSSLCaPath)
	setString("MIGRATIONS_PREFIX", &config.MigrationsPath)
	setString("CONTACT_ADDRESS", &config.ContactAddress)
	setString("LOGGING_FILENAME", &config.Logging.Filename)
	setString("LOGGING_LEVEL", &config.Logging.Level)
	return nil
}

func setString(name string, target *string) {
	if value, exists := os.LookupEnv(name); exists {
		*target = value
	}
}

func setBool(name string, target *bool) error {
	value, exists := os.LookupEnv(name)
	if !exists {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	*target = parsed
	return nil
}

func setCSV(name string, target *[]string) {
	value, exists := os.LookupEnv(name)
	if !exists {
		return
	}
	if value == "" {
		*target = []string{}
		return
	}
	parts := strings.Split(value, ",")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	*target = parts
}
