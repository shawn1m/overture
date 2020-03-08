package resolver

import "testing"

const (
	ipv4Address        = "8.8.8.8"
	ipv6Address        = "[2001:4860:4860::8888]"
	literalIpa6Address = "2001:4860:4860::8888"
)

func TestExtractDNSAddress(t *testing.T) {
	var tests = []struct {
		rawAddress string
		protocol   string
		host       string
		port       string
		err        error
	}{
		{"dns.google:853@" + ipv6Address, "tcp-tls", literalIpa6Address, "853", nil},
		{"dns.google:853@" + ipv4Address, "tcp-tls", ipv4Address, "853", nil},
		{ipv4Address + ":5353", "tcp", ipv4Address, "5353", nil},
		{ipv6Address + ":5353", "tcp", literalIpa6Address, "5353", nil},
		{ipv4Address + ":5353", "udp", ipv4Address, "5353", nil},
		{ipv6Address + ":5353", "udp", literalIpa6Address, "5353", nil},
		{ipv4Address, "udp", ipv4Address, "53", nil},
		{ipv6Address, "udp", literalIpa6Address, "53", nil},
		{"https://dns.google/dns-query", "https", "dns.google", "443", nil},
		{"https://dns.google:888/dns-query", "https", "dns.google", "888", nil},
	}
	for _, tt := range tests {
		t.Run(tt.rawAddress+", "+tt.protocol, func(t *testing.T) {
			host, port, err := ExtractDNSAddress(tt.rawAddress, tt.protocol)
			testEqual(t, host, tt.host)
			testEqual(t, port, tt.port)
			testErr(t, err)
		})
	}
}

func TestExtractSocksAddress(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{"socks5://" + ipv4Address + ":80", ipv4Address + ":80"},
		{"socks5://" + ipv6Address + ":80", ipv6Address + ":80"},
		{"socks5://" + ipv6Address, ipv6Address + ":1080"},
		{"" + ipv4Address + ":80", ipv4Address + ":80"},
		{"" + ipv6Address + ":80", ipv6Address + ":80"},
		{"" + ipv6Address, ipv6Address + ":1080"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			addr, err := ExtractSocksAddress(tt.in)
			testEqual(t, addr, tt.out)
			testErr(t, err)
		})
	}
}

func TestExtractTLSDNSAddress(t *testing.T) {

	var tests = []struct {
		in   string
		host string
		port string
		ip   string
		err  error
	}{
		{"dns.google:853@" + ipv6Address, "dns.google", "853", literalIpa6Address, nil},
		{"dns.google@" + ipv6Address, "dns.google", "853", literalIpa6Address, nil},
		{"dns.google:853@" + ipv4Address, "dns.google", "853", ipv4Address, nil},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			host, port, ip, err := ExtractTLSDNSAddress(tt.in)
			testEqual(t, host, tt.host)
			testEqual(t, port, tt.port)
			testEqual(t, ip, tt.ip)
			testErr(t, err)
		})
	}
}

func testEqual(t *testing.T, got string, want string) {
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func testErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("err is not nil: %s", err)
	}
}
