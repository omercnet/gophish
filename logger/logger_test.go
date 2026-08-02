package logger

import "testing"

import "github.com/sirupsen/logrus"

func TestSetup_uses_JSON_formatter_by_default(t *testing.T) {
	if err := Setup(&Config{}); err != nil {
		t.Fatalf("setup logger: %v", err)
	}
	if _, ok := Logger.Formatter.(*logrus.JSONFormatter); !ok {
		t.Fatalf("expected JSON formatter, got %T", Logger.Formatter)
	}
}

func TestSetup_uses_text_formatter_when_pretty(t *testing.T) {
	if err := Setup(&Config{Pretty: true}); err != nil {
		t.Fatalf("setup logger: %v", err)
	}
	if _, ok := Logger.Formatter.(*logrus.TextFormatter); !ok {
		t.Fatalf("expected text formatter, got %T", Logger.Formatter)
	}
}

func TestLogLevel(t *testing.T) {
	tests := map[string]logrus.Level{
		"":      logrus.InfoLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"error": logrus.ErrorLevel,
		"fatal": logrus.FatalLevel,
	}
	config := &Config{}
	for level, expected := range tests {
		config.Level = level
		err := Setup(config)
		if err != nil {
			t.Fatalf("error setting logging level %v", err)
		}
		if Logger.Level != expected {
			t.Fatalf("invalid logging level. expected %v got %v", expected, Logger.Level)
		}
	}
}
