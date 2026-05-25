package securityutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GenerateAESKey generates a 256-bit AES key.
func GenerateAESKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateIv generates a 16-byte initialization vector.
func GenerateIv() ([]byte, error) {
	iv := make([]byte, 16)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}
	return iv, nil
}

// Encrypt encrypts a string using AES/GCM/NoPadding (256-bit key, 16-byte IV, 128-bit tag).
func Encrypt(plainText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv, err := GenerateIv()
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, iv, []byte(plainText), nil)

	combined := make([]byte, len(iv)+len(ciphertext))
	copy(combined[0:16], iv)
	copy(combined[16:], ciphertext)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt decrypts a combined base64 GCM payload using the provided key.
func Decrypt(encryptedText string, key []byte) (string, error) {
	combined, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(combined) < 16 {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := combined[0:16]
	ciphertext := combined[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return "", err
	}

	plainBytes, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plainBytes), nil
}

// GetKeyFromString decodes a Base64 encoded key.
func GetKeyFromString(keyStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(keyStr)
}

// ConvertKeyToString encodes a key to Base64 string.
func ConvertKeyToString(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

// HashSHA256 returns SHA-256 hash in hex representation.
func HashSHA256(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

// HashSHA512 returns SHA-512 hash in hex representation.
func HashSHA512(input string) string {
	h := sha512.Sum512([]byte(input))
	return hex.EncodeToString(h[:])
}

// HashMD5 returns MD5 hash in hex representation.
func HashMD5(input string) string {
	h := md5.Sum([]byte(input))
	return hex.EncodeToString(h[:])
}

// GenerateSecureToken generates a 32-byte secure random URL-safe token.
func GenerateSecureToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}

// GenerateUUID generates standard UUID.
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateSecurePw generates a secure password of a given length.
func GenerateSecurePw(length int) (string, error) {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+"
	var pw strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		pw.WriteByte(chars[n.Int64()])
	}
	return pw.String(), nil
}

// GenerateHMAC generates HMAC-SHA256 hex string.
func GenerateHMAC(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSalt generates a 16-byte random salt base64 string.
func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// HashPw hashes password using bcrypt.
func HashPw(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPw checks if a password matches its bcrypt hash.
func VerifyPw(hashedPw, pw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPw), []byte(pw))
	return err == nil
}

// GenerateRandomNumber generates a random numeric string of given length.
func GenerateRandomNumber(length int) (string, error) {
	var num strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		num.WriteString(fmt.Sprintf("%d", n.Int64()))
	}
	return num.String(), nil
}
