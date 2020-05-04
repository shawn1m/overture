package common

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"net"
	"sort"
)

type ipRange struct {
	start net.IP
	end   net.IP
}

type ipRanges []*ipRange

func (s ipRanges) Len() int { return len(s) }
func (s ipRanges) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ipRanges) Less(i, j int) bool {
	return bytes.Compare(s[i].start, s[j].start) < 0
}

type IPSet struct {
	ipv4 ipRanges
	ipv6 ipRanges
}

func (s ipRanges) contains(ip net.IP) bool {
	l, r := 0, len(s)-1
	for l <= r {
		mid := (l + r) / 2
		if bytes.Compare(ip, s[mid].start) < 0 {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return r >= 0 && bytes.Compare(s[r].start, ip) <= 0 && bytes.Compare(ip, s[r].end) <= 0
}

func (ipSet *IPSet) Contains(ip net.IP, isLog bool, name string) bool {
	result := false
	if ipSet != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			result = ipSet.ipv4.contains(ipv4)
		}
		if !result && ipSet.ipv6 != nil {
			if ipv6 := ip.To16(); ipv6 != nil {
				result = ipSet.ipv6.contains(ipv6)
			}
		}
		if result && isLog {
			log.Debugf("Matched: IP network %s %s", name, ip.String())
		}
	} else {
		log.Debug("IP network list is nil, not checking")
	}
	return result
}

func toRange(ipNet *net.IPNet) *ipRange {
	ip, mask := ipNet.IP, ipNet.Mask
	ipLen := len(ip)
	start, end := make(net.IP, ipLen), make(net.IP, ipLen)
	for i := 0; i < ipLen; i++ {
		start[i] = ip[i] & mask[i]
		end[i] = ip[i] | ^mask[i]
	}
	return &ipRange{start, end}
}

func allFF(ip []byte) bool {
	for _, c := range ip {
		if c != 0xff {
			return false
		}
	}
	return true
}

func addOne(ip net.IP) net.IP {
	ipLen := len(ip)
	to := make(net.IP, ipLen)
	var carry uint = 1
	for i := ipLen - 1; i >= 0; i-- {
		carry += uint(ip[i])
		to[i] = byte(carry)
		carry >>= 8
	}
	return to
}

func sortAndMerge(rr ipRanges) ipRanges {
	if len(rr) < 2 {
		return rr
	}
	sort.Sort(rr)

	res := make(ipRanges, 0, len(rr))
	now := rr[0]
	start, end := now.start, now.end
	for i, count := 1, len(rr); i < count; i++ {
		now := rr[i]
		if allFF(end) || bytes.Compare(addOne(end), now.start) >= 0 {
			if bytes.Compare(end, now.end) < 0 {
				end = now.end
			}
		} else {
			res = append(res, &ipRange{start, end})
			start, end = now.start, now.end
		}
	}
	return append(res, &ipRange{start, end})
}

func NewIPSet(ipNetList []*net.IPNet) *IPSet {
	result := &IPSet{}
	for _, ipNet := range ipNetList {
		switch len(ipNet.IP) {
		case net.IPv4len:
			result.ipv4 = append(result.ipv4, toRange(ipNet))
		case net.IPv6len:
			result.ipv6 = append(result.ipv4, toRange(ipNet))
		default:
		}
	}
	if result.ipv4 == nil && result.ipv6 == nil {
		return nil
	}
	result.ipv6 = sortAndMerge(result.ipv6)
	result.ipv4 = sortAndMerge(result.ipv4)
	return result
}
