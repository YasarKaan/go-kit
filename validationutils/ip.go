package validationutils

import (
	"fmt"
	"net"
)

// IsValidIPv4 checks if the string is a valid IPv4 address.
func IsValidIPv4(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.To4() != nil
}

// IsValidIPv6 checks if the string is a valid IPv6 address.
func IsValidIPv6(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.To4() == nil
}

// IsValidIP checks if the string is a valid IPv4 or IPv6 address.
func IsValidIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}

// IsPrivateIP checks if the IP is in private ranges (including IPv4 RFC 1918 and IPv6 RFC 4193 ULA).
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ip.IsPrivate()
}

// IsLocalhost checks if the IP is localhost (127.0.0.1 or ::1).
func IsLocalhost(ipStr string) bool {
	return ipStr == "127.0.0.1" || ipStr == "::1" || IsLoopbackIP(ipStr)
}

// IsLoopbackIP checks if the IP is a loopback address.
func IsLoopbackIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.IsLoopback()
}

// IsLinkLocalIP checks if the IP is a link-local address.
func IsLinkLocalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast())
}

// IsSameSubnet checks if two IPv4 addresses are in the same subnet given a subnet mask.
func IsSameSubnet(ip1, ip2, subnetMask string) bool {
	if !IsValidIPv4(ip1) || !IsValidIPv4(ip2) || !IsValidIPv4(subnetMask) {
		return false
	}

	ip1Bytes := net.ParseIP(ip1).To4()
	ip2Bytes := net.ParseIP(ip2).To4()
	maskBytes := net.ParseIP(subnetMask).To4()

	if ip1Bytes == nil || ip2Bytes == nil || maskBytes == nil {
		return false
	}

	for i := 0; i < 4; i++ {
		if (ip1Bytes[i] & maskBytes[i]) != (ip2Bytes[i] & maskBytes[i]) {
			return false
		}
	}
	return true
}

// IpToLong converts an IPv4 address to a uint32/int64 representation.
func IpToLong(ipStr string) (int64, error) {
	if !IsValidIPv4(ipStr) {
		return 0, fmt.Errorf("invalid IPv4 address")
	}
	ipBytes := net.ParseIP(ipStr).To4()
	if ipBytes == nil {
		return 0, fmt.Errorf("invalid IPv4 address")
	}

	var result int64
	for i := 0; i < 4; i++ {
		result <<= 8
		result |= int64(ipBytes[i])
	}
	return result, nil
}

// LongToIp converts a long int to IPv4 address string.
func LongToIp(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(ip>>24)&0xFF,
		(ip>>16)&0xFF,
		(ip>>8)&0xFF,
		ip&0xFF)
}
