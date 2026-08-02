package imap

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestUIDSet_rejects_changed_UID_validity(t *testing.T) {
	_, err := uidSet([]Email{{UID: 42, UIDValidity: 10}}, 11)
	if !errors.Is(err, ErrMailboxRecreated) {
		t.Fatalf("expected mailbox recreation error, got %v", err)
	}
}

func TestUIDSet_uses_message_UIDs(t *testing.T) {
	set, err := uidSet([]Email{{UID: 42, UIDValidity: 10}}, 10)
	if err != nil {
		t.Fatalf("create UID set: %v", err)
	}
	if got := set.String(); got != "42" {
		t.Fatalf("expected UID 42, got %s", got)
	}
}

func TestParseMessage_rejects_oversized_source(t *testing.T) {
	_, err := parseMessage(bytes.NewReader(bytes.Repeat([]byte("a"), maxRawMessageBytes+1)))
	if !errors.Is(err, ErrMessageTooLarge) {
		t.Fatalf("expected message size error, got %v", err)
	}
}

func TestParseMessage_rejects_oversized_headers(t *testing.T) {
	raw := "X-Test: " + strings.Repeat("a", maxHeaderBytes) + "\r\n\r\nbody"
	_, err := parseMessage(strings.NewReader(raw))
	if !errors.Is(err, ErrMessageHeadersTooLarge) {
		t.Fatalf("expected header size error, got %v", err)
	}
}

func TestParseMessage_accepts_valid_message(t *testing.T) {
	raw := "From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: test\r\n\r\nbody"
	message, err := parseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	if message.Subject != "test" {
		t.Fatalf("expected subject test, got %q", message.Subject)
	}
}
