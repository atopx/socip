package core

import (
	"net"
)

var allFF = net.ParseIP("255.255.255.255").To4()

func DecToIpv4(dec int) net.IP {
	ipv4 := make(net.IP, net.IPv4len)
	ipv4[0] = byte(dec >> 24)
	ipv4[1] = byte(dec >> 16)
	ipv4[2] = byte(dec >> 8)
	ipv4[3] = byte(dec)
	return ipv4
}

func IPRange2CIDRs(start net.IP, end net.IP) (CIDRs []*net.IPNet) {
	maxLen := 32
	start = start.To4()
	end = end.To4()
	for cmp(start, end) <= 0 {
		l := 32
		for l > 0 {
			m := net.CIDRMask(l-1, maxLen)
			if cmp(start, first(start, m)) != 0 || cmp(last(start, m), end) > 0 {
				break
			}
			l--
		}
		CIDRs = append(CIDRs, &net.IPNet{IP: start, Mask: net.CIDRMask(l, maxLen)})
		start = last(start, net.CIDRMask(l, maxLen))
		if cmp(start, allFF) == 0 {
			break
		}
		start = next(start)
	}
	return CIDRs
}

func next(ip net.IP) net.IP {
	n := len(ip)
	out := make(net.IP, n)
	value := false
	for n > 0 {
		n--
		if value {
			out[n] = ip[n]
			continue
		}
		if ip[n] < 255 {
			out[n] = ip[n] + 1
			value = true
			continue
		}
		out[n] = 0
	}
	return out
}

func cmp(startIp net.IP, endIp net.IP) int {
	l := len(startIp)
	for i := 0; i < l; i++ {
		if startIp[i] == endIp[i] {
			continue
		}
		if startIp[i] < endIp[i] {
			return -1
		}
		return 1
	}
	return 0
}

func first(ip net.IP, mask net.IPMask) net.IP {
	return ip.Mask(mask)
}

func last(ip net.IP, mask net.IPMask) net.IP {
	n := len(ip)
	out := make(net.IP, n)
	for i := 0; i < n; i++ {
		out[i] = ip[i] | ^mask[i]
	}
	return out
}
