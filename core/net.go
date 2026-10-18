package core

import (
	"net"
	"text/template"
)

// netFuncs — Net
//
// Domain: net
//
// Check IP addresses and CIDR ranges: whether an address is valid, private, or
// inside a range.
func netFuncs() template.FuncMap {
	return template.FuncMap{
		"cidrContains": CIDRContains,
		"ipValid":      IPValid,
		"ipIsPrivate":  IPIsPrivate,
	}
}

// CIDRContains reports whether an IP address falls within a CIDR block.
// Returns `false` for malformed input rather than an error.
//
// Keywords: network, cidr, ip, subnet, contains, range, address
//
// Example:
//
//	{{ cidrContains "10.0.0.0/8" "10.4.2.7" }}  → true
func CIDRContains(cidr, ip string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return network.Contains(parsed)
}

// IPValid reports whether a string is a valid IPv4 or IPv6 address.
//
// Keywords: network, ip, address, validate, valid, ipv4, ipv6
//
// Example:
//
//	{{ ipValid "10.4.2.7" }}  → true
//	{{ ipValid "10.4.2" }}    → false
func IPValid(s string) bool {
	return net.ParseIP(s) != nil
}

// IPIsPrivate reports whether an IP address falls within a private (RFC 1918 /
// RFC 4193) range. Covers `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, and
// `fc00::/7`.
//
// Keywords: network, ip, address, private, internal, rfc1918, cidr
//
// Example:
//
//	{{ ipIsPrivate "192.168.1.20" }}  → true
func IPIsPrivate(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	private := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
	}
	for _, block := range private {
		_, network, err := net.ParseCIDR(block)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
