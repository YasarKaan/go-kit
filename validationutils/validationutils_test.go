package validationutils

import (
	"testing"
)

func TestStringValidation(t *testing.T) {
	if !IsValidEmail("test@example.com") {
		t.Error("expected test@example.com to be valid email")
	}
	if IsValidEmail("invalid-email") {
		t.Error("expected invalid-email to be invalid")
	}

	if !IsValidUrl("https://google.com") {
		t.Error("expected https://google.com to be valid URL")
	}
	if IsValidUrl("just-a-string") {
		t.Error("expected just-a-string to be invalid URL")
	}

	if !IsValidUrlStrict("http://example.com/path") {
		t.Error("expected http://example.com/path to be strict URL")
	}
	if IsValidUrlStrict("/relative/path") {
		t.Error("expected relative path to be invalid strict URL")
	}

	if !IsValidPhoneNumber("+905554443322") {
		t.Error("expected valid phone number")
	}

	// Luhn Algorithm (Mastercard/Visa test numbers)
	if !IsValidCreditCard("49927398716") {
		t.Error("expected valid Luhn credit card")
	}
	if IsValidCreditCard("49927398717") {
		t.Error("expected invalid Luhn credit card")
	}

	if !IsNumeric("123456") {
		t.Error("expected numeric")
	}
	if IsNumeric("123a45") {
		t.Error("expected non-numeric")
	}

	if !IsAlphabetic("Hello") {
		t.Error("expected alphabetic")
	}

	if !IsAlphanumeric("Hello123") {
		t.Error("expected alphanumeric")
	}

	// Pw strength
	if !IsStrongPw("Strong123!") {
		t.Error("expected strong password")
	}
	if IsStrongPw("weak") {
		t.Error("expected weak password")
	}

	if !IsValidFileName("my-file_name.txt") {
		t.Error("expected valid filename")
	}
	if IsValidFileName("../invalid/file.txt") {
		t.Error("expected invalid filename")
	}
}

func TestIPValidation(t *testing.T) {
	if !IsValidIPv4("192.168.1.1") {
		t.Error("expected valid IPv4")
	}
	if IsValidIPv4("256.0.0.1") {
		t.Error("expected invalid IPv4")
	}

	if !IsValidIPv6("::1") {
		t.Error("expected valid IPv6")
	}

	if !IsValidIP("1.1.1.1") || !IsValidIP("fe80::1") {
		t.Error("expected valid IP")
	}

	if !IsPrivateIP("10.0.0.5") || !IsPrivateIP("172.16.100.1") || !IsPrivateIP("192.168.1.50") {
		t.Error("expected private IPs")
	}
	if IsPrivateIP("8.8.8.8") {
		t.Error("expected public IP not to be private")
	}

	if !IsLocalhost("127.0.0.1") || !IsLocalhost("::1") {
		t.Error("expected localhost")
	}

	if !IsLoopbackIP("127.0.0.1") || !IsLoopbackIP("::1") {
		t.Error("expected loopback IP")
	}

	if !IsLinkLocalIP("169.254.1.1") || !IsLinkLocalIP("fe80::1") {
		t.Error("expected link-local IP")
	}

	if !IsSameSubnet("192.168.1.10", "192.168.1.20", "255.255.255.0") {
		t.Error("expected same subnet")
	}
	if IsSameSubnet("192.168.1.10", "192.168.2.20", "255.255.255.0") {
		t.Error("expected different subnets")
	}

	longVal, err := IpToLong("192.168.1.1")
	if err != nil {
		t.Fatalf("failed to convert IP to long: %v", err)
	}

	ipStr := LongToIp(longVal)
	if ipStr != "192.168.1.1" {
		t.Errorf("expected long IP conversion roundtrip to match, got: %s", ipStr)
	}
}
