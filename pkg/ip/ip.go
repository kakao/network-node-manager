package ip

import (
	"net"
	"strings"
)

func IsVaildIP(addr string) bool {
	return net.ParseIP(addr) != nil
}

func IsIPv4Addr(addr string) bool {
	if !IsVaildIP(addr) {
		return false
	}
	return strings.Count(addr, ":") < 2
}

func IsIPv6Addr(addr string) bool {
	if !IsVaildIP(addr) {
		return false
	}
	return strings.Count(addr, ":") >= 2
}
