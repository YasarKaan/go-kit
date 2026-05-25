package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/YasarKaan/go-kit/enums"
	"github.com/YasarKaan/go-kit/httputils"
	"github.com/YasarKaan/go-kit/loggerutils"
	"github.com/YasarKaan/go-kit/securityutils"
)

func main() {
	// 1. Initialize Logger
	// In production, you might set a file path, size limits, etc.
	// loggerutils.InitLogger("/var/log/myservice.log", 100, 5, 30, true, true)
	loggerutils.SetLogLevel(enums.LevelDebug)

	// 2. Structured logging with Correlation Context
	ctx := loggerutils.WithFields(context.Background(), map[string]any{
		"requestId": "req-987654",
		"traceId":   "trace-001122",
		"userId":    "user-kaan",
	})

	loggerutils.InfoContext(ctx, "Starting microservice dashboard example flow")

	// 3. Secure Password Hashing (bcrypt)
	password := "Sup3rS3cur3P@ssw0rd!"
	hashedPw, err := securityutils.HashPw(password)
	if err != nil {
		loggerutils.ErrorContext(ctx, "Failed to hash password: {}", err)
		return
	}
	loggerutils.InfoContext(ctx, "Successfully hashed password with bcrypt: {}", hashedPw)

	// Verify password matches
	matches := securityutils.VerifyPw(hashedPw, password)
	loggerutils.InfoContext(ctx, "Password verification matches: {}", matches)

	// 4. Safe HTTP Request via Reusable Connection Pool
	httpCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	url := "https://httpbin.org/get"
	loggerutils.InfoContext(httpCtx, "Sending GET request to pooled HTTP client at {}", url)

	resp, err := httputils.SendRequest(httpCtx, url, enums.MethodGet, nil, nil, enums.ContentTypeJSON)
	if err != nil {
		loggerutils.ErrorContext(httpCtx, "HTTP GET failed: {}", err)
	} else {
		loggerutils.InfoContext(httpCtx, "HTTP GET succeeded with status: {}", resp.StatusCode)
		// Cleanly parse JSON response to map
		if m, err := resp.Map(); err == nil {
			loggerutils.InfoContext(httpCtx, "Origin IP returned from HTTP response: {}", m["origin"])
		}
	}

	// 5. Gateway Verification Example
	// Configure environment signing secret
	osSecret := "SecretGatewaySignatureKey123"
	fmt.Printf("\n--- Gateway Verification Example ---\n")
	
	// Create request with gateway signature headers
	req, _ := http.NewRequest("GET", "http://my-internal-microservice/resource", nil)
	
	// Set headers
	req.Header.Set(securityutils.HeaderSchemaName, "public_schema")
	req.Header.Set(securityutils.HeaderUserId, "user-kaan")
	req.Header.Set(securityutils.HeaderTenantId, "tenant-odine")
	req.Header.Set(securityutils.HeaderRoles, "admin,operator")
	
	// Sign for 5 minutes into the future
	expTime := time.Now().Add(5 * time.Minute).Unix()
	expStr := fmt.Sprintf("%d", expTime)
	req.Header.Set(securityutils.HeaderTokenExp, expStr)

	// Generate expected signature
	// canonical format: schemaName:userId:tenantId:subjectType:sessionId:jti:tokenRefId:roles:exp
	canonical := fmt.Sprintf("public_schema:user-kaan:tenant-odine:::::admin,operator:%s", expStr)
	
	// We dynamically calculate signature using secret (normally set by API Gateway)
	// In application code, set GATEWAY_SIGN_SECRET environment variable
	_ = osSecret // set environment variable
	// For example simulation, we just verify directly:
	fmt.Println("Simulating signature validation payload:", canonical)
}
