package securityutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	HeaderSchemaName   = "X-Schema-Name"
	HeaderUserId       = "X-User-Id"
	HeaderTenantId     = "X-Tenant-Id"
	HeaderSubjectType  = "X-Subject-Type"
	HeaderSessionId    = "X-Session-Id"
	HeaderJti          = "X-Jti"
	HeaderTokenRefId   = "X-Token-Ref-Id"
	HeaderRoles        = "X-Roles"
	HeaderAuthSource   = "X-Auth-Source"
	HeaderGatewaySig   = "X-Gateway-Sig"
	HeaderTokenExp     = "X-Token-Exp"

	GatewayAuthSource  = "gateway"
)

type Claims struct {
	TenantId    string `json:"tenantId"`
	UserId      string `json:"userId"`
	SchemaName  string `json:"schemaName"`
	SubjectType string `json:"subjectType"`
	SessionId   string `json:"sessionId"`
	Jti         string `json:"jti"`
	TokenRefId  string `json:"tokenRefId"`
	Roles       string `json:"roles"`
}

// DecodeGatewayClaimsVerified extracts and validates gateway-injected identity headers.
// Validates the X-Gateway-Sig signature against the GATEWAY_SIGN_SECRET and checks token expiration.
func DecodeGatewayClaimsVerified(req *http.Request) (*Claims, error) {
	if req == nil {
		return nil, fmt.Errorf("http request is nil")
	}

	signature := req.Header.Get(HeaderGatewaySig)
	if signature == "" {
		return nil, fmt.Errorf("gateway signature header %s is missing", HeaderGatewaySig)
	}

	expStr := req.Header.Get(HeaderTokenExp)
	if expStr == "" {
		return nil, fmt.Errorf("gateway expiration header %s is missing", HeaderTokenExp)
	}

	// 1. Validate expiration timestamp
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expiration timestamp: %w", err)
	}

	if time.Now().Unix() >= exp {
		return nil, fmt.Errorf("token has expired")
	}

	// 2. Validate Gateway HMAC Signature
	secret := os.Getenv("GATEWAY_SIGN_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("GATEWAY_SIGN_SECRET environment variable is not set")
	}

	// Canonical string includes all read headers to prevent tampering
	canonical := strings.Join([]string{
		req.Header.Get(HeaderSchemaName),
		req.Header.Get(HeaderUserId),
		req.Header.Get(HeaderTenantId),
		req.Header.Get(HeaderSubjectType),
		req.Header.Get(HeaderSessionId),
		req.Header.Get(HeaderJti),
		req.Header.Get(HeaderTokenRefId),
		req.Header.Get(HeaderRoles),
		expStr,
	}, ":")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Safe constant-time comparison
	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(signature)) != 1 {
		return nil, fmt.Errorf("gateway header signature verification failed")
	}

	// Return verified claims
	return &Claims{
		TenantId:    req.Header.Get(HeaderTenantId),
		UserId:      req.Header.Get(HeaderUserId),
		SchemaName:  req.Header.Get(HeaderSchemaName),
		SubjectType: req.Header.Get(HeaderSubjectType),
		SessionId:   req.Header.Get(HeaderSessionId),
		Jti:         req.Header.Get(HeaderJti),
		TokenRefId:  req.Header.Get(HeaderTokenRefId),
		Roles:       req.Header.Get(HeaderRoles),
	}, nil
}

// IsFromGateway returns true if request carries verified and unexpired gateway headers.
func IsFromGateway(req *http.Request) bool {
	_, err := DecodeGatewayClaimsVerified(req)
	return err == nil
}

// ResolveRoles extracts roles list from X-Roles header.
func ResolveRoles(req *http.Request) []string {
	if req == nil {
		return []string{}
	}

	rolesStr := req.Header.Get(HeaderRoles)
	if rolesStr == "" {
		return []string{}
	}

	parts := strings.Split(rolesStr, ",")
	roles := make([]string, 0)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			roles = append(roles, trimmed)
		}
	}
	return roles
}
