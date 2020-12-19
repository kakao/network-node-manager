package ip

import (
	"fmt"
	"net"
	"strconv"
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

func GetAddrMaskFromCIDR(cidr string) (string, int, error) {
	if strings.Count(cidr, "/") != 1 {
		return "", 0, fmt.Errorf("wrong CIDR format")
	}

	tokens := strings.Split(cidr, "/")
	addr := tokens[0]
	mask := tokens[1]

	if !IsVaildIP(addr) {
		return "", 0, fmt.Errorf("wrong addr")
	}
	maskNum, err := strconv.Atoi(mask)
	if err != nil {
		return "", 0, fmt.Errorf("wrong mask")
	}
	return addr, maskNum, nil
}

func IsIPv4CIDR(cidr string) bool {
	addr, mask, err := GetAddrMaskFromCIDR(cidr)
	if err != nil {
		return false
	}

	if !IsIPv4Addr(addr) {
		return false
	}
	if mask > 32 {
		return false
	}
	return true
}

func IsIPv6CIDR(cidr string) bool {
	addr, mask, err := GetAddrMaskFromCIDR(cidr)
	if err != nil {
		return false
	}

	if !IsIPv6Addr(addr) {
		return false
	}
	if mask > 128 {
		return false
	}
	return true
}
