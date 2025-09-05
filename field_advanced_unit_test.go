// field_advanced_unit_test.go: Tests for advanced field types (ErrorField, NamedError, Stringer, Object, Errors)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Test types implementing fmt.Stringer for reuse across tests
type testStringer struct {
	value string
}

func (ts testStringer) String() string {
	return ts.value
}

type customStringer struct {
	id   int
	name string
}

func (cs customStringer) String() string {
	return fmt.Sprintf("CustomStringer{id=%d, name=%s}", cs.id, cs.name)
}

// User type for integration tests
type testUser struct {
	ID   int
	Name string
}

func (u testUser) String() string {
	return fmt.Sprintf("User{id=%d, name=%s}", u.ID, u.Name)
}

// TestErrorField tests ErrorField function
func TestErrorField(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected Field
	}{
		{
			name: "nil_error",
			err:  nil,
			expected: Field{
				K:   "error",
				T:   kindError,
				Obj: nil,
			},
		},
		{
			name: "simple_error",
			err:  errors.New("test error"),
			expected: Field{
				K:   "error",
				T:   kindError,
				Obj: errors.New("test error"),
			},
		},
		{
			name: "formatted_error",
			err:  fmt.Errorf("formatted error: %d", 404),
			expected: Field{
				K:   "error",
				T:   kindError,
				Obj: fmt.Errorf("formatted error: %d", 404),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorField(tt.err)

			// Check key and type
			if result.K != tt.expected.K {
				t.Errorf("Expected key '%s', got '%s'", tt.expected.K, result.K)
			}
			if result.T != tt.expected.T {
				t.Errorf("Expected type %v, got %v", tt.expected.T, result.T)
			}

			// Check error object
			if tt.err == nil {
				if result.Obj != nil {
					t.Error("Expected nil object for nil error")
				}
			} else {
				if result.Obj == nil {
					t.Error("Expected non-nil object for non-nil error")
				}
				if resultErr, ok := result.Obj.(error); ok {
					if resultErr.Error() != tt.err.Error() {
						t.Errorf("Expected error message '%s', got '%s'", tt.err.Error(), resultErr.Error())
					}
				} else {
					t.Error("Expected object to be an error")
				}
			}
		})
	}
}

// TestNamedError tests NamedError function
func TestNamedError(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		err      error
		expected Field
	}{
		{
			name: "custom_key_with_error",
			key:  "custom_error",
			err:  errors.New("custom error message"),
			expected: Field{
				K:   "custom_error",
				T:   kindError,
				Obj: errors.New("custom error message"),
			},
		},
		{
			name: "validation_error",
			key:  "validation",
			err:  fmt.Errorf("validation failed: field 'email' is required"),
			expected: Field{
				K:   "validation",
				T:   kindError,
				Obj: fmt.Errorf("validation failed: field 'email' is required"),
			},
		},
		{
			name: "nil_error_with_custom_key",
			key:  "auth_error",
			err:  nil,
			expected: Field{
				K:   "auth_error",
				T:   kindError,
				Obj: nil,
			},
		},
		{
			name: "empty_key",
			key:  "",
			err:  errors.New("error with empty key"),
			expected: Field{
				K:   "",
				T:   kindError,
				Obj: errors.New("error with empty key"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NamedError(tt.key, tt.err)

			// Check key and type
			if result.K != tt.expected.K {
				t.Errorf("Expected key '%s', got '%s'", tt.expected.K, result.K)
			}
			if result.T != tt.expected.T {
				t.Errorf("Expected type %v, got %v", tt.expected.T, result.T)
			}

			// Check error object
			if tt.err == nil {
				if result.Obj != nil {
					t.Error("Expected nil object for nil error")
				}
			} else {
				if result.Obj == nil {
					t.Error("Expected non-nil object for non-nil error")
				}
				if resultErr, ok := result.Obj.(error); ok {
					if resultErr.Error() != tt.err.Error() {
						t.Errorf("Expected error message '%s', got '%s'", tt.err.Error(), resultErr.Error())
					}
				} else {
					t.Error("Expected object to be an error")
				}
			}
		})
	}
}

// TestStringer tests Stringer function
func TestStringer(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		val      interface{ String() string }
		expected Field
	}{
		{
			name: "simple_stringer",
			key:  "test_string",
			val:  testStringer{value: "hello world"},
			expected: Field{
				K:   "test_string",
				T:   kindStringer,
				Obj: testStringer{value: "hello world"},
			},
		},
		{
			name: "custom_stringer",
			key:  "user",
			val:  customStringer{id: 123, name: "john"},
			expected: Field{
				K:   "user",
				T:   kindStringer,
				Obj: customStringer{id: 123, name: "john"},
			},
		},
		{
			name: "empty_key_stringer",
			key:  "",
			val:  testStringer{value: "empty key test"},
			expected: Field{
				K:   "",
				T:   kindStringer,
				Obj: testStringer{value: "empty key test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Stringer(tt.key, tt.val)

			// Check key and type
			if result.K != tt.expected.K {
				t.Errorf("Expected key '%s', got '%s'", tt.expected.K, result.K)
			}
			if result.T != tt.expected.T {
				t.Errorf("Expected type %v, got %v", tt.expected.T, result.T)
			}

			// Check object
			if result.Obj == nil {
				t.Error("Expected non-nil object")
			} else {
				// Verify the stringer interface works
				if stringer, ok := result.Obj.(interface{ String() string }); ok {
					expectedString := tt.val.String()
					resultString := stringer.String()
					if resultString != expectedString {
						t.Errorf("Expected string '%s', got '%s'", expectedString, resultString)
					}
				} else {
					t.Error("Expected object to implement String() method")
				}
			}
		})
	}
}

// TestObject tests Object function
func TestObject(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	tests := []struct {
		name     string
		key      string
		val      interface{}
		expected Field
	}{
		{
			name: "string_object",
			key:  "message",
			val:  "hello world",
			expected: Field{
				K:   "message",
				T:   kindObject,
				Obj: "hello world",
			},
		},
		{
			name: "struct_object",
			key:  "user",
			val:  testStruct{Name: "John", ID: 123},
			expected: Field{
				K:   "user",
				T:   kindObject,
				Obj: testStruct{Name: "John", ID: 123},
			},
		},
		{
			name: "map_object",
			key:  "config",
			val:  map[string]interface{}{"enabled": true, "timeout": 30},
			expected: Field{
				K:   "config",
				T:   kindObject,
				Obj: map[string]interface{}{"enabled": true, "timeout": 30},
			},
		},
		{
			name: "slice_object",
			key:  "items",
			val:  []string{"apple", "banana", "cherry"},
			expected: Field{
				K:   "items",
				T:   kindObject,
				Obj: []string{"apple", "banana", "cherry"},
			},
		},
		{
			name: "nil_object",
			key:  "null_value",
			val:  nil,
			expected: Field{
				K:   "null_value",
				T:   kindObject,
				Obj: nil,
			},
		},
		{
			name: "pointer_object",
			key:  "ptr",
			val:  &testStruct{Name: "Jane", ID: 456},
			expected: Field{
				K:   "ptr",
				T:   kindObject,
				Obj: &testStruct{Name: "Jane", ID: 456},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Object(tt.key, tt.val)

			// Check key and type
			if result.K != tt.expected.K {
				t.Errorf("Expected key '%s', got '%s'", tt.expected.K, result.K)
			}
			if result.T != tt.expected.T {
				t.Errorf("Expected type %v, got %v", tt.expected.T, result.T)
			}

			// Check object using reflection for deep comparison
			if !reflect.DeepEqual(result.Obj, tt.expected.Obj) {
				t.Errorf("Expected object %+v, got %+v", tt.expected.Obj, result.Obj)
			}
		})
	}
}

// TestErrors tests Errors function
func TestErrors(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		errs     []error
		expected Field
	}{
		{
			name: "multiple_errors",
			key:  "validation_errors",
			errs: []error{
				errors.New("field 'name' is required"),
				errors.New("field 'email' is invalid"),
				errors.New("field 'age' must be positive"),
			},
			expected: Field{
				K: "validation_errors",
				T: kindObject,
				Obj: []error{
					errors.New("field 'name' is required"),
					errors.New("field 'email' is invalid"),
					errors.New("field 'age' must be positive"),
				},
			},
		},
		{
			name: "single_error",
			key:  "error",
			errs: []error{errors.New("single error")},
			expected: Field{
				K:   "error",
				T:   kindObject,
				Obj: []error{errors.New("single error")},
			},
		},
		{
			name: "empty_errors",
			key:  "no_errors",
			errs: []error{},
			expected: Field{
				K:   "no_errors",
				T:   kindObject,
				Obj: []error{},
			},
		},
		{
			name: "nil_errors",
			key:  "nil_errors",
			errs: nil,
			expected: Field{
				K:   "nil_errors",
				T:   kindObject,
				Obj: ([]error)(nil),
			},
		},
		{
			name: "errors_with_nil_elements",
			key:  "mixed_errors",
			errs: []error{
				errors.New("real error"),
				nil,
				fmt.Errorf("formatted error: %d", 500),
			},
			expected: Field{
				K: "mixed_errors",
				T: kindObject,
				Obj: []error{
					errors.New("real error"),
					nil,
					fmt.Errorf("formatted error: %d", 500),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Errors(tt.key, tt.errs)

			// Check key and type
			if result.K != tt.expected.K {
				t.Errorf("Expected key '%s', got '%s'", tt.expected.K, result.K)
			}
			if result.T != tt.expected.T {
				t.Errorf("Expected type %v, got %v", tt.expected.T, result.T)
			}

			// Check errors slice
			resultErrs, ok := result.Obj.([]error)
			if !ok {
				t.Error("Expected object to be []error")
				return
			}

			if tt.errs == nil {
				if resultErrs != nil {
					t.Error("Expected nil slice for nil errors")
				}
			} else {
				if len(resultErrs) != len(tt.errs) {
					t.Errorf("Expected %d errors, got %d", len(tt.errs), len(resultErrs))
					return
				}

				for i, expectedErr := range tt.errs {
					if expectedErr == nil {
						if resultErrs[i] != nil {
							t.Errorf("Expected nil error at index %d, got %v", i, resultErrs[i])
						}
					} else {
						if resultErrs[i] == nil {
							t.Errorf("Expected error at index %d, got nil", i)
						} else if resultErrs[i].Error() != expectedErr.Error() {
							t.Errorf("Expected error '%s' at index %d, got '%s'",
								expectedErr.Error(), i, resultErrs[i].Error())
						}
					}
				}
			}
		})
	}
}

// bufferedTestSyncer for capturing output in tests
type fieldTestSyncer struct {
	data []byte
}

func (bs *fieldTestSyncer) Write(p []byte) (n int, err error) {
	bs.data = append(bs.data, p...)
	return len(p), nil
}

func (bs *fieldTestSyncer) Sync() error {
	return nil
}

func (bs *fieldTestSyncer) String() string {
	return string(bs.data)
}

// TestAdvancedFieldsIntegration tests advanced fields in actual logging scenarios
func TestAdvancedFieldsIntegration(t *testing.T) {
	buf := &fieldTestSyncer{}
	logger, err := New(Config{
		Level:    Debug,
		Encoder:  NewJSONEncoder(),
		Output:   buf,
		Capacity: 1024, // Safe capacity for CI
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Start()
	defer safeCloseFieldLogger(t, logger)

	// Log with various advanced field types
	testUser := testUser{ID: 123, Name: "John Doe"}
	testErrors := []error{
		errors.New("validation error"),
		fmt.Errorf("network error: timeout"),
	}

	logger.Info("testing advanced fields",
		ErrorField(errors.New("main error")),
		NamedError("auth_error", fmt.Errorf("unauthorized: %s", "invalid token")),
		Stringer("user", testUser),
		Object("config", map[string]interface{}{"debug": true, "timeout": 30}),
		Errors("validation_errors", testErrors),
	)

	// Allow async processing
	time.Sleep(50 * time.Millisecond)
	_ = logger.Sync()

	output := buf.String()

	// Verify the log contains expected elements
	expectedPatterns := []string{
		"testing advanced fields",
		"error",
		"main error",
		"auth_error",
		"unauthorized",
		"user",
		"config",
		"validation_errors",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected output to contain '%s', got: %s", pattern, output)
		}
	}

	// Verify it's valid JSON
	if !isValidJSONFieldTest(output) {
		t.Errorf("Output should be valid JSON, got: %s", output)
	}
}

// Helper functions for testing

func safeCloseFieldLogger(t *testing.T, logger *Logger) {
	if err := logger.Close(); err != nil {
		t.Logf("Warning: Error closing logger in test: %v", err)
	}
}

func isValidJSONFieldTest(s string) bool {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(line), &js); err != nil {
			return false
		}
	}
	return true
}
