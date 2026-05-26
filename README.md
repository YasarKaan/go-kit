# go-kit

> An opinionated, lightweight microservice utility toolkit for Go.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Overview

`go-kit` is a collection of cohesive, independent subpackages designed to simplify common microservices boilerplate in Go: structured JSON logging, connection-pooled HTTP clients, gateway validation, input checks, and date arithmetic.

**Design Principles:**
- **Context-First** — Every HTTP request and logging call propagates `context.Context` for cancellation and distributed tracing.
- **Secure Defaults** — Bcrypt password hashing, AES-256-GCM encryption, CRLF-injection checks, and safe constant-time HMAC validations.
- **Connection Reuse** — Reuses a configured global connection pool (transports) to prevent socket/port exhaustion under load.

---

## Packages

| Package | Purpose |
|---|---|
| [`loggerutils`](#loggerutils) | Structured JSON logging with sensitive field masking and trace context |
| [`httputils`](#httputils) | Pooled HTTP client with idempotent retries and file stream uploading |
| [`securityutils`](#securityutils) | Bcrypt hashing, AES-256-GCM encryption, and Gateway validation |
| [`stringutils`](#stringutils) | Type-safe parameter map extraction and JSON generics |
| [`validationutils`](#validationutils) | Email, phone, IP subnet (`net.IPNet`), and filename checking |
| [`dateutils`](#dateutils) | Weekdays/business days arithmetic and strict ISO8601 formatting |
| [`fileutils`](#fileutils) | Memory-safe streaming files processing and OOM protection |
| [`exceptions`](#exceptions) | Custom structured HTTP error models |
| [`enums`](#enums) | Common HTTP method, log level, and priority constants |

---

## Installation

```bash
go get github.com/YasarKaan/go-kit
```

---

## Reference Guides

### `loggerutils`
Structured JSON logging featuring sensitive-data masking and context trace fields:

```go
import (
    "context"
    "github.com/YasarKaan/go-kit/loggerutils"
)

// Masking is automatic for strings and nested JSON (e.g. password, secret, token keys)
loggerutils.Info("Login attempt: password={}", "s3cr3t")

// Tracing / Correlation context
ctx := loggerutils.WithFields(context.Background(), map[string]any{
    "requestId": "req-12345",
    "userId":    "user-kaan",
})
loggerutils.InfoContext(ctx, "Processing payments")
// Output: {"time":"...","level":"INFO","message":"Processing payments","requestId":"req-12345","userId":"user-kaan"}

// Instance-based logging is also fully supported for project encapsulation:
logger := loggerutils.NewLogger(os.Stdout, enums.LevelDebug)
logger.Info("This is an isolated instance")
```

---

### `httputils`
HTTP client built on a reusable, tuned transport connection pool (`MaxIdleConnsPerHost = 100`) to avoid socket exhaustion.

- **Safe Retries**: Automatic retries are restricted to **idempotent methods** (`GET`, `HEAD`, `PUT`, `DELETE`). Non-idempotent methods (`POST`, `PATCH`) are **not retried** unless an `Idempotency-Key` header is explicitly provided.
- **Timeout**: Default client timeout is set to a microservice-appropriate `30` seconds.

```go
// Standard Pooled request
resp, err := httputils.SendRequest(ctx, "https://api.service.com/users", enums.MethodGet, nil, nil, enums.ContentTypeJSON)

// Request with safe retries (respects context cancellation in retry wait)
resp, err := httputils.SendRequestWithRetries(ctx, "https://api.service.com/checkout", enums.MethodPost, map[string]string{"Idempotency-Key": "key-123"}, payload)

// SSL-skipping helpers are explicitly prefixed as dangerous
resp, err := httputils.DangerousSendRequestWithoutSSL(ctx, "https://internal/dev", enums.MethodGet, nil, nil, enums.ContentTypeJSON)
```

---

### `securityutils`
Bcrypt password hashing, AES-256-GCM encryption, and gateway signature verification.

- **Password Hashing (bcrypt)**:
```go
hashed, err := securityutils.HashPw("userPassword123!")
isValid := securityutils.VerifyPw(hashed, "userPassword123!") // true
```

- **AES-256-GCM Encryption** (Standard 12-byte nonce Base64 combined):
```go
key, _ := securityutils.GenerateAESKey()
encrypted, _ := securityutils.Encrypt("sensitive info", key)
decrypted, _ := securityutils.Decrypt(encrypted, key)
```

- **Gateway Verification**:
Verify pre-authenticated headers signed by your API Gateway via HMAC-SHA256 and unix timestamp expiration:
```go
// Expose X-Gateway-Sig validations. Canonical string signs all read identity headers.
claims, err := securityutils.DecodeGatewayClaimsVerified(r)
if err == nil {
    fmt.Printf("Authenticated User: %s, Schema: %s\n", claims.UserId, claims.SchemaName)
}
```

---

### `stringutils`
Type-safe parameter map extraction and JSON serialization/conversion helpers.

- **Parameter Extraction (`variable.go`)**:
Extract values from generic maps (query, body, or path params) safely. It supports returning default values, pointers, or custom web exceptions (HTTP 422) automatically when fields are missing or invalid:

```go
import "github.com/YasarKaan/go-kit/stringutils"

// Merge parameters from different sources
params := stringutils.GetParameterMap(queryParams, bodyParams, pathParams)

// Safe extraction with fallback defaults
user := stringutils.GetStringValueFromMapWithDefault(params, "username", "guest")
limit := stringutils.GetLongValueFromMapWithDefault(params, "limit", 10)

// Extraction with automatic HTTP 422 error wrapping if value is invalid/missing
uuidVal, err := stringutils.GetUUIDValueFromMapWithException(params, "userId", "userId is required and must be a valid UUID")
```

- **JSON Helper & Jackson-like Conversion (`json.go`)**:
```go
// Type-safe JSON parsing into generics
userObj, err := stringutils.JsonStringToObject[User](`{"name":"kaan"}`)

// Deep copy / Type convert (similar to Jackson's ObjectMapper.convertValue in Java)
var target User
target, err = stringutils.ConvertValue[User](map[string]any{"name": "kaan"})
```

---

### `validationutils`
Validation and parsing helpers for strings, formats, and IP addresses.

- **String Format Validation (`string.go`)**:
Verify formatting for common schemas, including strict URL parsing, password strength, and credit card validation (Luhn algorithm):
```go
import "github.com/YasarKaan/go-kit/validationutils"

// General string validations
isValid := validationutils.IsValidEmail("kaan@example.com")
isStrong := validationutils.IsStrongPw("SecurePass123!") // Length >= 8, mixed cases, digit, special char
isValidCard := validationutils.IsValidCreditCard("49927398716") // Luhn algorithm check

// URL and files
isStrictURL := validationutils.IsValidUrlStrict("https://example.com")
isSafeFile := validationutils.IsValidFileName("invoice_2026.pdf")
```

- **IP Parsing & Utilities (`ip.go`)**:
```go
// Check IP versions or private subnets
isIPv4 := validationutils.IsValidIPv4("192.168.1.1")
isPrivate := validationutils.IsPrivateIP("10.0.0.5") // Supports RFC 1918 and RFC 4193 (IPv6 ULA)
isLoopback := validationutils.IsLoopbackIP("127.0.0.1") // true
isLinkLocal := validationutils.IsLinkLocalIP("169.254.0.1") // true
isSameNet := validationutils.IsSameSubnet("192.168.1.10", "192.168.1.20", "255.255.255.0")

// Integer/Long conversion (for DB indexes or ranges)
ipLong, _ := validationutils.IpToLong("192.168.1.1") // 3232235777
ipStr := validationutils.LongToIp(ipLong) // "192.168.1.1"
```

---

### `dateutils`
Date format parsing, calculations, and business day counters using native Go/ISO8601 layout standards with strict error propagation.

- **Date Formatting & Range Math**:
Parse and process using native Go time formats (e.g. `"2006-01-02"`):
```go
import "github.com/YasarKaan/go-kit/dateutils"

// Parse custom formats using native Go layouts
isValid := dateutils.IsValid("2026-05-26 12:00:00", "2006-01-02 15:04:05")

// Add days or compute relative time (all return errors on parsing failure)
newDate, err := dateutils.AddDays("2026-05-26", "2006-01-02", 5) // "2026-05-31", nil
daysDiff, err := dateutils.GetDaysBetween("2026-05-01", "2026-05-10", "2006-01-02", true) // 10, nil
age, err := dateutils.CalculateAge("1995-10-15", "2006-01-02")
```

- **Business Day Utilities**:
```go
// Checks if a date falls on a weekday
isWorkday, err := dateutils.IsBusinessDay("2026-05-24", "2006-01-02") // false, nil (Sunday)

// Count weekdays between range
workdays, err := dateutils.CountBusinessDays("2026-05-25", "2026-05-29", "2006-01-02", true) // 5, nil (Mon-Fri)
```

---

### `fileutils`
Avoids loading large file buffers into memory by using streaming transfers.

```go
// OOM-Safe Streaming Multipart file
file, err := fileutils.FromFileStream("/tmp/large-log.txt", "text/plain")

// Write directly using io.Copy streaming (no memory allocation)
err = file.SaveToFile("/data/backups/log.txt")
```

---

### `exceptions`
Structured exceptions mapping to API JSON error payloads. Automatically redacts `Authorization`, `Cookie`, and `Set-Cookie` headers, and masks passwords in body parameters when attached to error metadata maps.

```go
return exceptions.NewCustomWebServerException(422, "Invalid order", map[string]any{
    "headers": r.Header,
    "payload": orderBody, // Redacted & masked automatically in metadata logs
})
```

---

## Running Tests

Verify everything runs successfully:
```bash
go test -race ./...
```
