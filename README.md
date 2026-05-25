# go-kit

`go-kit` is a lightweight, Go-idiomatic utility library designed to simplify common backend tasks. It provides safe type conversions, HTTP request execution with automatic backoff retries, logging sanitization, cryptographic helper functions, business day calculations, and distributed tracing/audit logging.

This library is built with simplicity in mind (KISS) and is structured into cohesive, independent subpackages.

---

## Features & Subpackages

### 1. `stringutils` (JSON & Map Utilities)
Safely extract typed values from generic `map[string]any` configurations or JSON request bodies with default fallbacks, wrapping `github.com/spf13/cast`.
```go
import "github.com/YasarKaan/go-kit/stringutils"

data := map[string]any{"status": "RUNNING", "timeout": 30}

// Safe extraction with fallback defaults
status := stringutils.GetStringValueFromMapWithDefault(data, "status", "PENDING") // "RUNNING"
limit := stringutils.GetIntegerValueFromMapWithDefault(data, "limit", 10)         // 10

// Convert dynamic payloads (Jackson-like ConvertValue)
type Config struct { Name string }
cfg, err := stringutils.ConvertValue[Config](map[string]any{"Name": "FunProject"})
```

### 2. `validationutils` (String & IP Validations)
Validate emails, phone numbers, password strength, Luhn algorithms (credit cards), IPv4/IPv6, private subnets, loopbacks, and files names.
```go
import "github.com/YasarKaan/go-kit/validationutils"

isValidEmail := validationutils.IsValidEmail("test@example.com")
isPrivateIP := validationutils.IsPrivateIP("192.168.1.100")
isSameSubnet := validationutils.IsSameSubnet("192.168.1.10", "192.168.1.20", "255.255.255.0")
```

### 3. `loggerutils` (Sensitive Data Masking Logger)
Protects against Log Forging (CRLF injection) and automatically masks sensitive fields (such as `password`, `token`, `secret`, `api_key`, JWT strings, etc.) to `***` within logging payloads.
```go
import "github.com/YasarKaan/go-kit/loggerutils"

// Automatically sanitizes CRLFs and masks sensitive variables/JSON
loggerutils.Info("User logged in with password: {}", "mySecretPassword123") // masks to "mySecretPassword123" -> "***"
```

### 4. `dateutils` (Business Day & Calendar Calculations)
Provides date formatting, relative offsets, age calculations, quarters, and business day (weekdays Monday-Friday) counting. Includes an built-in translator converting Java-style layout strings (e.g., `yyyy-MM-dd`) to Go layout strings.
```go
import "github.com/YasarKaan/go-kit/dateutils"

// Check format validity
isValid := dateutils.IsValid("2026-05-25", "yyyy-MM-dd")

// Count weekdays (inclusive of end date)
weekdays := dateutils.CountBusinessDays("2026-05-18", "2026-05-22", "yyyy-MM-dd", true) // 5 days
```

### 5. `httputils` (HTTP Request & Multipart Client)
Executes HTTP actions (GET, POST, PUT, DELETE) with support for URL-encoded forms, raw text, JSON payloads, and multipart form uploads. Includes automatic exponential backoff retries with jitter on server or rate-limiting errors.
```go
import (
	"github.com/YasarKaan/go-kit/enums"
	"github.com/YasarKaan/go-kit/httputils"
)

// Request with auto-retries on transient failures (429, 5xx)
response, err := httputils.SendRequestWithRetries("https://api.example.com/data", enums.MethodPost, nil, myPayload)
if err == nil {
	println(response.Body)
}

// Or request returning parsed JSON response directly as map[string]any
responseMap, err := httputils.SendRequestWithRetriesForMap("https://api.example.com/data", enums.MethodPost, nil, myPayload)
if err == nil {
	println(responseMap["status"])
}
```

### 6. `securityutils` (Crypto, Hashes, & JWT claims)
AES-256 GCM encryption/decryption (compatible with Java GCM 16-byte IVs), standard SHA/MD5 hashing, HMAC-SHA256 generation, random tokens/passwords, and unverified JWT payload extraction.
```go
import "github.com/YasarKaan/go-kit/securityutils"

// AES GCM with 16-byte nonce
key, _ := securityutils.GenerateAESKey()
encrypted, _ := securityutils.Encrypt("plainTextData", key)
decrypted, _ := securityutils.Decrypt(encrypted, key)

// Gateway Signature verification
isVerified := securityutils.IsFromGateway(httpRequest)
```

### 7. `logingestorclient` (RabbitMQ Audit Trails)
A context-aware RabbitMQ client helper to log session flows and mission steps for distributed auditing.
```go
import "github.com/YasarKaan/go-kit/logingestorclient"

// Initialize queue connection
logingestorclient.Initialize("localhost", 5672, "audit-exchange", "audit-queue", "audit-route")

// Start a mission step, returns UUID and updated context
missionId, newCtx := logingestorclient.StartTheMission(ctx, "ImportAction", "Importing templates", 1)

// Finalize mission step
logingestorclient.FinalizeTheMission(newCtx, "operation-uuid-123", "Bearer auth-token")
```

---

## Installation

```bash
go get github.com/YasarKaan/go-kit
```

## Running Tests

Verify everything runs successfully:
```bash
go test ./...
```
