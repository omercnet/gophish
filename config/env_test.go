package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	log "github.com/gophish/gophish/logger"
)

var configEnvVars = []string{
	"ADMIN_LISTEN_URL",
	"ADMIN_USE_TLS",
	"ADMIN_CERT_PATH",
	"ADMIN_KEY_PATH",
	"ADMIN_CSRF_KEY",
	"ADMIN_ALLOWED_INTERNAL_HOSTS",
	"ADMIN_TRUSTED_ORIGINS",
	"PHISH_LISTEN_URL",
	"PHISH_USE_TLS",
	"PHISH_CERT_PATH",
	"PHISH_KEY_PATH",
	"DB_NAME",
	"DB_PATH",
	"DB_FILE_PATH",
	"DB_SSLCA_PATH",
	"MIGRATIONS_PREFIX",
	"CONTACT_ADDRESS",
	"LOGGING_FILENAME",
	"LOGGING_LEVEL",
	"LOG_PRETTY",
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range configEnvVars {
		value, exists := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("unset %s: %v", name, err)
		}
		t.Cleanup(func() {
			if exists {
				if err := os.Setenv(name, value); err != nil {
					t.Errorf("restore %s: %v", name, err)
				}
				return
			}
			if err := os.Unsetenv(name); err != nil {
				t.Errorf("clear %s: %v", name, err)
			}
		})
	}
}

func TestLoadConfig_from_environment_without_file(t *testing.T) {
	clearConfigEnv(t)
	values := map[string]string{
		"ADMIN_LISTEN_URL":             "0.0.0.0:3333",
		"ADMIN_USE_TLS":                "true",
		"ADMIN_CERT_PATH":              "admin.crt",
		"ADMIN_KEY_PATH":               "admin.key",
		"ADMIN_CSRF_KEY":               "secret",
		"ADMIN_ALLOWED_INTERNAL_HOSTS": "10.0.0.0/8, 192.168.1.1",
		"ADMIN_TRUSTED_ORIGINS":        "https://admin.example, https://ops.example",
		"PHISH_LISTEN_URL":             "0.0.0.0:8080",
		"PHISH_USE_TLS":                "false",
		"PHISH_CERT_PATH":              "phish.crt",
		"PHISH_KEY_PATH":               "phish.key",
		"DB_NAME":                      "sqlite3",
		"DB_PATH":                      "env.db",
		"DB_SSLCA_PATH":                "ca.pem",
		"MIGRATIONS_PREFIX":            "migrations/",
		"CONTACT_ADDRESS":              "security@example.com",
		"LOGGING_FILENAME":             "gophish.log",
		"LOGGING_LEVEL":                "debug",
		"LOG_PRETTY":                   "true",
	}
	for name, value := range values {
		t.Setenv(name, value)
	}

	conf, err := LoadConfig("")
	if err != nil {
		t.Fatalf("load environment config: %v", err)
	}

	expected := &Config{
		AdminConf: AdminServer{
			ListenURL:            "0.0.0.0:3333",
			UseTLS:               true,
			CertPath:             "admin.crt",
			KeyPath:              "admin.key",
			CSRFKey:              "secret",
			AllowedInternalHosts: []string{"10.0.0.0/8", "192.168.1.1"},
			TrustedOrigins:       []string{"https://admin.example", "https://ops.example"},
		},
		PhishConf: PhishServer{
			ListenURL: "0.0.0.0:8080",
			UseTLS:    false,
			CertPath:  "phish.crt",
			KeyPath:   "phish.key",
		},
		DBName:         "sqlite3",
		DBPath:         "env.db",
		DBSSLCaPath:    "ca.pem",
		MigrationsPath: "migrations/sqlite3",
		ContactAddress: "security@example.com",
		Logging:        &log.Config{Filename: "gophish.log", Level: "debug", Pretty: true},
	}
	if !reflect.DeepEqual(expected, conf) {
		t.Fatalf("invalid environment config. expected %#v got %#v", expected, conf)
	}
}

func TestLoadConfig_environment_overrides_file(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, validConfig, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("ADMIN_LISTEN_URL", "127.0.0.1:4444")
	t.Setenv("DB_PATH", "override.db")

	conf, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if conf.AdminConf.ListenURL != "127.0.0.1:4444" || conf.DBPath != "override.db" {
		t.Fatalf("environment did not override file: %#v", conf)
	}
}

func TestLoadConfig_rejects_invalid_boolean_environment_value(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("ADMIN_USE_TLS", "sometimes")

	_, err := LoadConfig("")
	if err == nil || !strings.Contains(err.Error(), "ADMIN_USE_TLS") {
		t.Fatalf("expected ADMIN_USE_TLS parse error, got %v", err)
	}
}

func TestLoadConfig_supports_legacy_DB_FILE_PATH(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DB_FILE_PATH", "legacy.db")

	conf, err := LoadConfig("")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if conf.DBPath != "legacy.db" {
		t.Fatalf("expected legacy.db, got %q", conf.DBPath)
	}
}
