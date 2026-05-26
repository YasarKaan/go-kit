package securityutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestAESEncryptionDecryption(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("failed to generate AES key: %v", err)
	}

	plain := "Secret message 123!"
	encrypted, err := Encrypt(plain, key)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if decrypted != plain {
		t.Errorf("expected decrypted text to match plain text, got: %s", decrypted)
	}
}

func TestHashingAndHMAC(t *testing.T) {
	input := "hello world"
	sha256Hex := HashSHA256(input)
	if len(sha256Hex) != 64 {
		t.Error("expected 64 character hex string for sha256")
	}

	sha512Hex := HashSHA512(input)
	if len(sha512Hex) != 128 {
		t.Error("expected 128 character hex string for sha512")
	}

	md5Hex := HashMD5Insecure(input)
	if len(md5Hex) != 32 {
		t.Error("expected 32 character hex string for md5")
	}

	hmacHex := GenerateHMAC(input, "secretKey")
	if len(hmacHex) != 64 {
		t.Error("expected 64 character hex string for hmac-sha256")
	}
}

func TestGenerateTokens(t *testing.T) {
	tok, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("unexpected error generating secure token: %v", err)
	}
	if len(tok) == 0 {
		t.Error("expected non-empty secure token")
	}

	pwd, err := GenerateSecurePw(12)
	if err != nil {
		t.Fatalf("unexpected error generating secure password: %v", err)
	}
	if len(pwd) != 12 {
		t.Errorf("expected 12 character password, got: %d", len(pwd))
	}

	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("unexpected error generating salt: %v", err)
	}
	if len(salt) == 0 {
		t.Error("expected non-empty salt")
	}

	num, err := GenerateRandomNumber(6)
	if err != nil {
		t.Fatalf("unexpected error generating random number: %v", err)
	}
	if len(num) != 6 {
		t.Errorf("expected random number of length 6, got: %d", len(num))
	}
}

func TestBcryptPasswordHashing(t *testing.T) {
	pw := "SecretPassword123!"
	hashed, err := HashPw(pw)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !VerifyPw(hashed, pw) {
		t.Error("expected password verification to succeed")
	}

	if VerifyPw(hashed, "wrong_password") {
		t.Error("expected password verification to fail for wrong password")
	}
}

func TestGatewayVerificationAndExpiry(t *testing.T) {
	secret := "secretKey123"
	os.Setenv("GATEWAY_SIGN_SECRET", secret)
	defer os.Unsetenv("GATEWAY_SIGN_SECRET")

	expTime := time.Now().Add(5 * time.Minute).Unix()
	expStr := fmt.Sprintf("%d", expTime)

	// canonical: schemaName:userId:tenantId:subjectType:sessionId:jti:tokenRefId:roles:exp
	canonical := fmt.Sprintf("s1:u1:t1:sub1:sess1:jti1:ref1:admin,user:%s", expStr)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(HeaderSchemaName, "s1")
	req.Header.Set(HeaderUserId, "u1")
	req.Header.Set(HeaderTenantId, "t1")
	req.Header.Set(HeaderSubjectType, "sub1")
	req.Header.Set(HeaderSessionId, "sess1")
	req.Header.Set(HeaderJti, "jti1")
	req.Header.Set(HeaderTokenRefId, "ref1")
	req.Header.Set(HeaderRoles, "admin,user")
	req.Header.Set(HeaderTokenExp, expStr)
	req.Header.Set(HeaderGatewaySig, expectedSig)

	// Test IsFromGateway
	if !IsFromGateway(req) {
		t.Error("expected gateway signature verification to succeed")
	}

	// Test DecodeGatewayClaimsVerified
	claims, err := DecodeGatewayClaimsVerified(req)
	if err != nil {
		t.Fatalf("unexpected error decoding claims: %v", err)
	}

	if claims.SchemaName != "s1" || claims.UserId != "u1" || claims.TenantId != "t1" || claims.Roles != "admin,user" {
		t.Errorf("decoded claims do not match: %+v", claims)
	}

	// Test ResolveRoles
	roles := ResolveRoles(req)
	if len(roles) != 2 || roles[0] != "admin" || roles[1] != "user" {
		t.Errorf("unexpected roles: %v", roles)
	}

	// Test Expiration Verification (set exp in past)
	pastExpStr := fmt.Sprintf("%d", time.Now().Add(-5*time.Minute).Unix())
	pastCanonical := fmt.Sprintf("s1:u1:t1:sub1:sess1:jti1:ref1:admin,user:%s", pastExpStr)
	macPast := hmac.New(sha256.New, []byte(secret))
	macPast.Write([]byte(pastCanonical))
	pastSig := base64.StdEncoding.EncodeToString(macPast.Sum(nil))

	reqPast, _ := http.NewRequest("GET", "http://example.com", nil)
	reqPast.Header.Set(HeaderSchemaName, "s1")
	reqPast.Header.Set(HeaderUserId, "u1")
	reqPast.Header.Set(HeaderTenantId, "t1")
	reqPast.Header.Set(HeaderSubjectType, "sub1")
	reqPast.Header.Set(HeaderSessionId, "sess1")
	reqPast.Header.Set(HeaderJti, "jti1")
	reqPast.Header.Set(HeaderTokenRefId, "ref1")
	reqPast.Header.Set(HeaderRoles, "admin,user")
	reqPast.Header.Set(HeaderTokenExp, pastExpStr)
	reqPast.Header.Set(HeaderGatewaySig, pastSig)

	if IsFromGateway(reqPast) {
		t.Error("expected gateway verification to fail for expired token")
	}

	_, err = DecodeGatewayClaimsVerified(reqPast)
	if err == nil || err.Error() != "token has expired" {
		t.Errorf("expected expired error, got: %v", err)
	}
}
