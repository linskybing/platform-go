package config

import (
	"os"
	"testing"
)

func TestParseEnvInt64(t *testing.T) {
	setEnv(t, "TEST_INT64", "123")
	if got := parseEnvInt64("TEST_INT64", 42); got != 123 {
		t.Fatalf("expected 123, got %d", got)
	}

	setEnv(t, "TEST_INT64", "bad")
	if got := parseEnvInt64("TEST_INT64", 42); got != 42 {
		t.Fatalf("expected fallback 42, got %d", got)
	}
}

func TestGetEnv(t *testing.T) {
	setEnv(t, "TEST_ENV", "value")
	if got := getEnv("TEST_ENV", "fallback"); got != "value" {
		t.Fatalf("expected value, got %s", got)
	}
	if got := getEnv("MISSING_ENV", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestLoadConfigQueueSettings(t *testing.T) {
	setEnv(t, "GO_ENV", "development")
	setEnv(t, "CONFIGFILE_QUEUE_PRIORITY", "200")
	setEnv(t, "DEFAULT_QUEUE_PRIORITY", "50")
	setEnv(t, "CONFIGFILE_QUEUE_MAX_CONCURRENT", "10")
	setEnv(t, "DEFAULT_QUEUE_MAX_CONCURRENT", "12")
	setEnv(t, "CONFIGFILE_QUEUE_TTL_SECONDS", "30")
	setEnv(t, "DEFAULT_QUEUE_TTL_SECONDS", "60")
	setEnv(t, "FLASH_SCHED_QUEUE_ANNOTATION_KEY", "queue-key")
	setEnv(t, "FLASH_SCHED_PREEMPTABLE_ANNOTATION_KEY", "preempt-key")
	setEnv(t, "CONFIGFILE_PREEMPTABLE", "true")

	LoadConfig()

	if ConfigFileQueuePriority != 200 {
		t.Fatalf("expected ConfigFileQueuePriority 200, got %d", ConfigFileQueuePriority)
	}
	if DefaultQueuePriority != 50 {
		t.Fatalf("expected DefaultQueuePriority 50, got %d", DefaultQueuePriority)
	}
	if ConfigFileQueueMaxConcurrent != 10 {
		t.Fatalf("expected ConfigFileQueueMaxConcurrent 10, got %d", ConfigFileQueueMaxConcurrent)
	}
	if DefaultQueueMaxConcurrent != 12 {
		t.Fatalf("expected DefaultQueueMaxConcurrent 12, got %d", DefaultQueueMaxConcurrent)
	}
	if ConfigFileQueueTTLSeconds != 30 {
		t.Fatalf("expected ConfigFileQueueTTLSeconds 30, got %d", ConfigFileQueueTTLSeconds)
	}
	if DefaultQueueTTLSeconds != 60 {
		t.Fatalf("expected DefaultQueueTTLSeconds 60, got %d", DefaultQueueTTLSeconds)
	}
	if FlashSchedQueueAnnotationKey != "queue-key" {
		t.Fatalf("expected FlashSchedQueueAnnotationKey queue-key, got %s", FlashSchedQueueAnnotationKey)
	}
	if FlashSchedPreemptableAnnotationKey != "preempt-key" {
		t.Fatalf("expected FlashSchedPreemptableAnnotationKey preempt-key, got %s", FlashSchedPreemptableAnnotationKey)
	}
	if !ConfigFilePreemptable {
		t.Fatalf("expected ConfigFilePreemptable true")
	}
}

func setEnv(t *testing.T, key, value string) {
	old, ok := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("setenv %s failed: %v", key, err)
	}
	t.Cleanup(func() {
		if !ok {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, old)
	})
}
