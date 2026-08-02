//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
)

func TestMailpit_receives_generated_email(t *testing.T) {
	t.Setenv(models.InitialAdminPassword, "integration-test-password")
	if err := models.Setup(&config.Config{
		DBName:         "sqlite3",
		DBPath:         filepath.Join(t.TempDir(), "gophish.db"),
		MigrationsPath: "../db/db_sqlite3/migrations/",
	}); err != nil {
		t.Fatalf("setup models: %v", err)
	}

	smtpAddress := envOrDefault("MAILPIT_SMTP_ADDR", "127.0.0.1:1025")
	request := &models.EmailRequest{
		Template:      models.Template{Subject: "Mailpit integration", Text: "Hello {{.FirstName}}", HTML: "<p>Hello {{.FirstName}}</p>"},
		SMTP:          models.SMTP{Host: smtpAddress, FromAddress: "sender@example.com"},
		URL:           "http://example.com",
		RId:           "preview-mailpit",
		FromAddress:   "sender@example.com",
		BaseRecipient: models.BaseRecipient{Email: "recipient@example.com", FirstName: "Integration", LastName: "Test"},
		ErrorChan:     make(chan error, 1),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mailWorker := mailer.NewMailWorker()
	go mailWorker.Start(ctx)
	mailWorker.Queue([]mailer.Mail{request})
	if err := <-request.ErrorChan; err != nil {
		t.Fatalf("send email: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(envOrDefault("MAILPIT_API_URL", "http://127.0.0.1:8025") + "/api/v1/messages?start=0&limit=50")
	if err != nil {
		t.Fatalf("query Mailpit: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Mailpit returned %s", response.Status)
	}
	var mailbox struct {
		Total    int `json:"total"`
		Messages []struct {
			Subject string `json:"subject"`
			Snippet string `json:"snippet"`
			From    struct {
				Address string `json:"address"`
			} `json:"from"`
			To []struct {
				Address string `json:"address"`
			} `json:"to"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(response.Body).Decode(&mailbox); err != nil {
		t.Fatalf("decode Mailpit response: %v", err)
	}
	if mailbox.Total != 1 || len(mailbox.Messages) != 1 {
		t.Fatalf("expected one message, got %d", mailbox.Total)
	}
	message := mailbox.Messages[0]
	if message.Subject != "Mailpit integration" || message.From.Address != "sender@example.com" || len(message.To) != 1 || message.To[0].Address != "recipient@example.com" {
		t.Fatalf("unexpected message: %+v", message)
	}
	if !strings.Contains(message.Snippet, "Hello Integration") {
		t.Fatalf("unexpected snippet %q", message.Snippet)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
