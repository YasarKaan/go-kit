package securityutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"testing"
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

	md5Hex := HashMD5(input)
	if len(md5Hex) != 32 {
		t.Error("expected 32 character hex string for md5")
	}

	hmacHex := GenerateHMAC(input, "secretKey")
	if len(hmacHex) != 64 {
		t.Error("expected 64 character hex string for hmac-sha256")
	}
}

func TestGenerateTokens(t *testing.T) {
	tok := GenerateSecureToken()
	if len(tok) == 0 {
		t.Error("expected non-empty secure token")
	}

	pwd := GenerateSecurePw(12)
	if len(pwd) != 12 {
		t.Errorf("expected 12 character password, got: %d", len(pwd))
	}
}

func TestJWTDecode(t *testing.T) {
	// Sample JWT payload: {"tenantId": "t1", "sub": "u1", "schemaName": "s1", "sid": "sess1"}
	payload := `{"tenantId": "t1", "sub": "u1", "schemaName": "s1", "sid": "sess1"}`
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	token := "header." + encodedPayload + ".signature"

	claims := DecodeToken(token)
	if claims == nil {
		t.Fatal("expected claims to be decoded")
	}

	if claims.TenantId != "t1" || claims.UserId != "u1" || claims.SchemaName != "s1" || claims.SessionId != "sess1" {
		t.Errorf("decoded claims do not match: %+v", claims)
	}
}

func TestGatewayVerification(t *testing.T) {
	secret := "secretKey123"
	os.Setenv("GATEWAY_SIGN_SECRET", secret)
	defer os.Unsetenv("GATEWAY_SIGN_SECRET")

	// Calculate HMAC of "s1:u1:t1:1780000000"
	canonical := "s1:u1:t1:1780000000"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(HeaderSchemaName, "s1")
	req.Header.Set(HeaderUserId, "u1")
	req.Header.Set(HeaderTenantId, "t1")
	req.Header.Set(HeaderTokenExp, "1780000000")
	req.Header.Set(HeaderGatewaySig, expectedSig)
	req.Header.Set(HeaderAuthSource, GatewayAuthSource)

	if !IsFromGateway(req) {
		t.Error("expected gateway signature verification to succeed")
	}

	claims := DecodeTokenFromRequest(req)
	if claims == nil || claims.SchemaName != "s1" || claims.UserId != "u1" {
		t.Errorf("failed to decode claims from request, got: %+v", claims)
	}
}
