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
- **No JWT Fallback Traps** — Legacy JWT decoding is removed. Instead, services verify pre-authenticated gateway signature headers.

---

## Packages

| Package | Purpose |
|---|---|
| [`loggerutils`](#loggerutils) | Structured JSON logging with sensitive field masking and trace context |
| [`httputils`](#httputils) | Pooled HTTP client with idempotent retries and file stream uploading |
| [`securityutils`](#securityutils) | Bcrypt hashing, AES-256-GCM encryption, and Gateway validation |
| [`stringutils`](#stringutils) | Type-safe parameter map extraction and JSON generics |
| [`validationutils`](#validationutils) | Email, phone, IP subnet (`net.IPNet`), and filename checking |
| [`dateutils`](#dateutils) | Weekdays/business days math with Java formatting layout translation |
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

- **AES-256-GCM Encryption** (16-byte nonce Base64 combined - Java compatible):
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
