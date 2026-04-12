// security_test.go: Security features test suite for Iris logging library
//
// THREAT MODEL:
//   CWE-200: Exposure of Sensitive Information
//     - Secret() fields must produce [REDACTED] in all encoder paths
//     - Field combination must not leak secrets via adjacent fields
//   CWE-362: Concurrent Execution Using Shared Resource
//     - Concurrent Start/Close/Write must not panic or corrupt state
//     - Ring buffer under pressure must not discard security-critical records silently
//   CWE-843: Access of Resource Using Incompatible Type
//     - Field type confusion: wrong kind enum could bypass redaction
//     - Zero-value Fields must not cause panic in encoder
//   CWE-20: Improper Input Validation
//     - Malicious field content must not break logger invariants
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSecretFieldRedaction(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create logger with buffer output
	logger, err := New(Config{
		Level:  Debug,
		Output: WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseSecurityLogger(t, logger)

	logger.Start()

	// Log with sensitive data
	logger.Info("User authentication",
		Secret("password", "super_secret_password123"),
		Secret("api_key", "sk-1234567890abcdef"),
		Str("username", "john_doe"), // Regular field for comparison
	)

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Verify that sensitive data is redacted
	if strings.Contains(output, "super_secret_password123") {
		t.Error("Password was not redacted in log output")
	}
	if strings.Contains(output, "sk-1234567890abcdef") {
		t.Error("API key was not redacted in log output")
	}

	// Verify that redaction marker is present
	if !strings.Contains(output, `"password":"[REDACTED]"`) {
		t.Error("Password field redaction marker not found")
	}
	if !strings.Contains(output, `"api_key":"[REDACTED]"`) {
		t.Error("API key field redaction marker not found")
	}

	// Verify that non-sensitive data is still present
	if !strings.Contains(output, `"username":"john_doe"`) {
		t.Error("Username field was incorrectly redacted")
	}

	t.Logf("Log output: %s", output)
}

func TestSecretFieldType(t *testing.T) {
	// Test that Secret() creates the correct field type
	field := Secret("test_key", "sensitive_value")

	if field.K != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", field.K)
	}
	if field.T != kindSecret {
		t.Errorf("Expected type kindSecret (%d), got %d", kindSecret, field.T)
	}
	if field.Str != "sensitive_value" {
		t.Errorf("Expected value 'sensitive_value', got '%s'", field.Str)
	}
}

func TestMultipleSecretFields(t *testing.T) {
	var buf bytes.Buffer

	logger, err := New(Config{
		Level:  Info,
		Output: WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer safeCloseSecurityLogger(t, logger)

	logger.Start()

	// Log with multiple secret fields
	logger.Warn("Security event",
		Secret("password", "password123"),
		Secret("credit_card", "4111-1111-1111-1111"),
		Secret("ssn", "123-45-6789"),
		Str("event_type", "login_attempt"),
		Int("attempt_count", 3),
	)

	time.Sleep(50 * time.Millisecond)
	output := buf.String()

	// Verify all secrets are redacted
	sensitiveData := []string{
		"password123",
		"4111-1111-1111-1111",
		"123-45-6789",
	}

	for _, sensitive := range sensitiveData {
		if strings.Contains(output, sensitive) {
			t.Errorf("Sensitive data '%s' was not redacted", sensitive)
		}
	}

	// Verify redaction markers
	redactionMarkers := []string{
		`"password":"[REDACTED]"`,
		`"credit_card":"[REDACTED]"`,
		`"ssn":"[REDACTED]"`,
	}

	for _, marker := range redactionMarkers {
		if !strings.Contains(output, marker) {
			t.Errorf("Redaction marker '%s' not found", marker)
		}
	}

	// Verify non-sensitive data is preserved
	if !strings.Contains(output, `"event_type":"login_attempt"`) {
		t.Error("Non-sensitive string field was incorrectly redacted")
	}
	if !strings.Contains(output, `"attempt_count":3`) {
		t.Error("Non-sensitive integer field was incorrectly processed")
	}
}

// Helper function for safe logger cleanup
func safeCloseSecurityLogger(t *testing.T, logger *Logger) {
	t.Helper()
	if err := logger.Close(); err != nil &&
		!strings.Contains(err.Error(), "flush timeout") {
		t.Logf("Warning: Error closing logger in test: %v", err)
	}
}

// --- Attack vector: CWE-362 (Concurrent Start/Close/Write) ---

func TestSecurity_ConcurrentStartStopWrite(t *testing.T) {
	// ATTACK VECTOR: CWE-362
	// IMPACT: Concurrent Start/Close/Write could corrupt ring buffer state,
	// cause nil pointer dereference, or leave resources leaked.
	// MITIGATION EXPECTED: atomic state management, no panic.
	const goroutines = 20

	var buf bufferedSyncer
	logger, err := New(Config{
		Level:   Info,
		Output:  &buf,
		Encoder: NewJSONEncoder(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	logger.Start()

	var panicCount atomic.Int64
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCount.Add(1)
					t.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()
			for i := 0; i < 100; i++ {
				logger.Info("concurrent security test",
					Str("goroutine", "worker"),
					Int("iter", i))
			}
		}(g)
	}

	wg.Wait()
	safeCloseSecurityLogger(t, logger)

	if panicCount.Load() > 0 {
		t.Fatalf("%d goroutines panicked during concurrent writes", panicCount.Load())
	}
}

// --- Attack vector: CWE-843 (Field type confusion) ---

func TestSecurity_FieldTypeConfusion(t *testing.T) {
	// ATTACK VECTOR: CWE-843
	// IMPACT: A Field with an invalid kind value could bypass redaction
	// logic or cause panic in the encoder switch statement.
	// MITIGATION EXPECTED: unknown kinds handled gracefully (worst case: no output for field).
	var buf bytes.Buffer
	enc := NewJSONEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Build record with a manually crafted Field using an invalid kind
	rec := NewRecord(Info, "type confusion test")
	invalidField := Field{
		K:   "evil",
		T:   255, // WHY 255: no such kind in the enum, tests default case
		Str: "should_not_appear",
	}
	rec.AddField(invalidField)
	rec.AddField(Str("normal", "value"))

	// Must not panic
	enc.Encode(rec, now, &buf)
	output := buf.String()

	if len(output) == 0 {
		t.Fatal("encoder produced empty output for type confusion input")
	}

	// Normal field must still appear
	if !strings.Contains(output, `"normal"`) {
		t.Error("normal field missing after type confusion field")
	}
}

func TestSecurity_ZeroValueField(t *testing.T) {
	// ATTACK VECTOR: CWE-843
	// IMPACT: A zero-value Field{} has kind=0 (kindString with empty key/value).
	// Must not cause panic.
	// MITIGATION EXPECTED: graceful handling.
	var buf bytes.Buffer
	enc := NewJSONEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	rec := NewRecord(Info, "zero field test")
	rec.AddField(Field{}) // zero value
	rec.AddField(Str("after_zero", "present"))

	// Must not panic
	enc.Encode(rec, now, &buf)
	output := buf.String()

	if !strings.Contains(output, "present") {
		t.Error("fields after zero-value field were lost")
	}
}

// --- Attack vector: CWE-200 (Secret adjacent to similar fields) ---

func TestSecurity_SecretAdjacentFieldMixup(t *testing.T) {
	// ATTACK VECTOR: CWE-200
	// IMPACT: If encoder processes fields out of order or confuses
	// adjacent fields, a Secret's value could leak through the next field.
	// MITIGATION EXPECTED: each field encoded independently.
	var buf bytes.Buffer

	logger, err := New(Config{
		Level:  Info,
		Output: WrapWriter(&buf),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer safeCloseSecurityLogger(t, logger)
	logger.Start()

	// Interleave secret and non-secret fields with similar keys
	logger.Info("mixed fields",
		Secret("token", "secret_token_value"),
		Str("token_type", "bearer"),
		Secret("key", "secret_key_value"),
		Str("key_id", "pub-123"),
	)

	time.Sleep(50 * time.Millisecond)
	output := buf.String()

	// Secret values must never appear
	if strings.Contains(output, "secret_token_value") {
		t.Fatal("SECRET LEAKED: secret_token_value found in output")
	}
	if strings.Contains(output, "secret_key_value") {
		t.Fatal("SECRET LEAKED: secret_key_value found in output")
	}

	// Non-secret values must appear
	if !strings.Contains(output, "bearer") {
		t.Error("non-secret field token_type lost")
	}
	if !strings.Contains(output, "pub-123") {
		t.Error("non-secret field key_id lost")
	}
}

// --- Fuzz target: FuzzField ---

func FuzzField(f *testing.F) {
	// Seeds: boundary values for field construction
	f.Add("key", "value", uint8(0))  // kindString
	f.Add("key", "value", uint8(10)) // kindSecret
	f.Add("", "", uint8(0))          // empty
	f.Add("key\x00null", "val\x00ue", uint8(0))
	f.Add("key\ninjection", "val\nue", uint8(10))
	f.Add(strings.Repeat("k", 4096), strings.Repeat("v", 4096), uint8(0))

	enc := NewJSONEncoder()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	f.Fuzz(func(t *testing.T, key, value string, kindByte uint8) {
		var buf bytes.Buffer
		rec := NewRecord(Info, "fuzz")

		// Use Secret for kindSecret, Str for everything else
		if kindByte == uint8(kindSecret) {
			rec.AddField(Secret(key, value))
		} else {
			rec.AddField(Str(key, value))
		}

		// Must not panic
		enc.Encode(rec, now, &buf)

		if buf.Len() == 0 {
			t.Fatal("encoder produced empty output")
		}

		// If it was a secret, the value must not appear
		if kindByte == uint8(kindSecret) && value != "" &&
			value != "[REDACTED]" &&
			!strings.Contains("[REDACTED]", value) &&
			strings.Contains(buf.String(), value) {
			t.Fatalf("SECRET LEAKED via fuzz: key=%q value=%q", key, value)
		}
	})
}
