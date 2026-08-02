package models

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gophish/gophish/config"
	log "github.com/gophish/gophish/logger"
)

func TestSetup_does_not_log_configured_initial_admin_password(t *testing.T) {
	const password = "operator-managed-secret"
	t.Setenv(InitialAdminPassword, password)
	previousOutput := log.Logger.Out
	var output bytes.Buffer
	log.Logger.SetOutput(&output)
	t.Cleanup(func() {
		log.Logger.SetOutput(previousOutput)
	})

	err := Setup(&config.Config{
		DBName:         "sqlite3",
		DBPath:         filepath.Join(t.TempDir(), "gophish.db"),
		MigrationsPath: "../db/db_sqlite3/migrations/",
	})
	if err != nil {
		t.Fatalf("setup database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	if strings.Contains(output.String(), password) {
		t.Fatal("configured initial admin password was written to logs")
	}
}
