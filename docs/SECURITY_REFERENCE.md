# Security Reference - Quick Guide

## Sensitive Data Masking

### Basic Usage
```go
// Automatically redacted in all outputs
logger.Info("Authentication",
    Str("user", "john_doe"),        // ✅ Visible
    Secret("password", "secret123"), // ❌ Hidden as [REDACTED]
    Secret("api_key", "sk-abc123"),  // ❌ Hidden as [REDACTED]
)
```

### Common Sensitive Data Types
```go
// Authentication
Secret("password", userPassword)
Secret("token", authToken)
Secret("session_id", sessionID)

// API & Integration
Secret("api_key", apiKey)
Secret("webhook_secret", webhookSecret)
Secret("connection_string", dbConnectionString)

// Financial
Secret("credit_card", creditCardNumber)
Secret("bank_account", accountNumber)
Secret("ssn", socialSecurityNumber)

// Personal Data
Secret("email", userEmail)           // If email is considered PII
Secret("phone", phoneNumber)         // If phone is considered PII
Secret("address", homeAddress)       // If address is considered PII
```

## Log Injection Protection

### Automatic Protection
All user input is automatically sanitized:

```go
// These are all safe - IRIS prevents injection automatically
logger.Info("User input",
    Str("filename", userFilename),           // Dangerous chars sanitized
    Str("search_query", userSearchQuery),    // Newlines become underscores  
    Str("comment", userComment),             // Quotes properly escaped
)
```

### Attack Examples (All Prevented)

#### Field Key Injection
```go
// Malicious input: `user" admin="true`
// Result: user__admin__true="normal_value"
logger.Info("Action", Str(maliciousKey, "normal_value"))
```

#### Field Value Injection  
```go
// Malicious input: `data\nlevel=error msg="HACKED"`
// Result: data="data_level_error msg_\"HACKED\""
logger.Info("Input", Str("data", maliciousValue))
```

#### Unicode Attacks
```go
// Malicious input: `user\u202e\u202d`  
// Result: user__="value"
logger.Info("Unicode", Str(maliciousUnicode, "value"))
```

## Performance Impact

```
Feature                 Overhead    Allocation
Normal Field           2.33 ns     0 allocs
Secret Field           2.38 ns     0 allocs  (+4.9%)
JSON Encoding          423 ns      0 allocs
Text Encoding          481 ns      0 allocs
```

## Security Guarantees

- ✅ **Zero sensitive data leakage**: Secret fields never appear in logs
- ✅ **Complete injection protection**: All input sanitized automatically  
- ✅ **Zero configuration**: Security enabled by default
- ✅ **Zero allocations**: No performance penalty for security
- ✅ **All encoders**: Works with JSON, Text, and future encoders

## Quick Integration

### Step 1: Replace sensitive fields
```go
// Before
Str("password", password)

// After  
Secret("password", password)
```

### Step 2: Use structured logging
```go
// Before (vulnerable)
logger.Printf("Error: %s", userInput)

// After (protected)
logger.Error("Error occurred", Str("details", userInput))
```

### Step 3: Test with malicious input
```go
// Test with injection attempts
testInput := "normal\nlevel=error msg=\"INJECTED\""
logger.Info("Test", Str("input", testInput))
// Verify output is properly sanitized
```

## Common Patterns

### Database Operations
```go
logger.Info("Database query",
    Str("operation", "SELECT"),
    Str("table", tableName),
    Secret("connection_string", dbURL),
    Int("duration_ms", queryTime),
)
```

### API Requests
```go
logger.Info("API call",
    Str("method", "POST"),
    Str("endpoint", "/api/users"),
    Secret("auth_header", authHeader),
    Int("status_code", response.StatusCode),
)
```

### User Actions
```go
logger.Info("User action",
    Str("action", "file_upload"),
    Str("filename", filename),        // Auto-sanitized
    Secret("session_token", token),
    Int64("file_size", fileSize),
)
```

### Error Handling
```go
logger.Error("Authentication failed",
    Str("username", username),
    Str("reason", "invalid_credentials"),
    Secret("attempted_password", password),
    Str("client_ip", clientIP),
)
```
