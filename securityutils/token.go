package securityutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/YasarKaan/go-kit/loggerutils"
)

const (
	HeaderSchemaName = "X-Schema-Name"
	HeaderUserId     = "X-User-Id"
	HeaderTenantId   = "X-Tenant-Id"
	HeaderSubjectType = "X-Subject-Type"
	HeaderSessionId  = "X-Session-Id"
	HeaderJti        = "X-Jti"
	HeaderTokenRefId = "X-Token-Ref-Id"
	HeaderTokenExp   = "X-Token-Exp"
	HeaderRoles      = "X-Roles"
	HeaderAuthSource = "X-Auth-Source"
	HeaderGatewaySig = "X-Gateway-Sig"

	GatewayAuthSource = "gateway"
)

type Claims struct {
	TenantId    string `json:"tenantId"`
	UserId      string `json:"sub"`
	SchemaName  string `json:"schemaName"`
	SubjectType string `json:"subjectType"`
	SessionId   string `json:"sid"`
	Jti         string `json:"jti"`
	TokenRefId  string `json:"tokenRefId"`
}

// DecodeToken decodes the payload of a JWT token without validating the signature (unverified decode).
func DecodeToken(token string) *Claims {
	if token == "" {
		return nil
	}

	rawToken := token
	if strings.HasPrefix(token, "Bearer ") {
		rawToken = token[7:]
	}

	parts := strings.Split(rawToken, ".")
	if len(parts) < 2 {
		loggerutils.Warn("JWT token does not have at least 2 parts")
		return nil
	}

	payload := parts[1]
	// Add padding to base64url if missing
	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}

	decodedBytes, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		loggerutils.Warn("JWT decode failed: %v", err)
		return nil
	}

	var claimsRaw map[string]any
	err = json.Unmarshal(decodedBytes, &claimsRaw)
	if err != nil {
		loggerutils.Warn("JWT payload JSON unmarshal failed: %v", err)
		return nil
	}

	// Handle both camelCase and snake_case or standard fields
	claims := &Claims{}
	if val, ok := claimsRaw["tenantId"].(string); ok {
		claims.TenantId = val
	}
	if val, ok := claimsRaw["sub"].(string); ok {
		claims.UserId = val
	}
	if val, ok := claimsRaw["schemaName"].(string); ok {
		claims.SchemaName = val
	} else if val, ok := claimsRaw["tenantSchemaName"].(string); ok {
		claims.SchemaName = val
	}
	if val, ok := claimsRaw["subjectType"].(string); ok {
		claims.SubjectType = val
	}
	if val, ok := claimsRaw["sid"].(string); ok {
		claims.SessionId = val
	}
	if val, ok := claimsRaw["jti"].(string); ok {
		claims.Jti = val
	}
	if val, ok := claimsRaw["tokenRefId"].(string); ok {
		claims.TokenRefId = val
	} else if val, ok := claimsRaw["token_ref_id"].(string); ok {
		claims.TokenRefId = val
	}

	return claims
}

// DecodeTokenFromRequest resolves claims from gateway headers or fallback to Authorization header.
func DecodeTokenFromRequest(req *http.Request) *Claims {
	if req == nil {
		return nil
	}

	schemaName := req.Header.Get(HeaderSchemaName)
	signature := req.Header.Get(HeaderGatewaySig)

	if schemaName != "" && signature != "" {
		claims := buildClaimsFromHeaders(req)
		if verifyGatewaySignature(claims, req.Header.Get(HeaderTokenExp), signature) {
			return claims
		}
		loggerutils.Warn("Gateway header signature verification failed")
	}

	return DecodeToken(req.Header.Get("Authorization"))
}

// IsFromGateway returns true if request carries verified gateway headers.
func IsFromGateway(req *http.Request) bool {
	if req == nil {
		return false
	}

	signature := req.Header.Get(HeaderGatewaySig)
	authSource := req.Header.Get(HeaderAuthSource)

	if signature == "" || authSource != GatewayAuthSource {
		return false
	}

	claims := buildClaimsFromHeaders(req)
	return verifyGatewaySignature(claims, req.Header.Get(HeaderTokenExp), signature)
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

func buildClaimsFromHeaders(req *http.Request) *Claims {
	return &Claims{
		TenantId:    req.Header.Get(HeaderTenantId),
		UserId:      req.Header.Get(HeaderUserId),
		SchemaName:  req.Header.Get(HeaderSchemaName),
		SubjectType: req.Header.Get(HeaderSubjectType),
		SessionId:   req.Header.Get(HeaderSessionId),
		Jti:         req.Header.Get(HeaderJti),
		TokenRefId:  req.Header.Get(HeaderTokenRefId),
	}
}

func verifyGatewaySignature(claims *Claims, exp string, signature string) bool {
	if claims == nil || signature == "" {
		return false
	}

	secret := os.Getenv("GATEWAY_SIGN_SECRET")
	if secret == "" {
		return false
	}

	canonical := strings.Join([]string{
		claims.SchemaName,
		claims.UserId,
		claims.TenantId,
		exp,
	}, ":")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Safe constant time comparison
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}
