package resolver

import "testing"

const (
	Ipv4Address        = "8.8.8.8"
	Ipv6Address        = "[2001:4860:4860::8888]"
	LiteralIpa6Address = "2001:4860:4860::8888"
)

func TestExtractDNSAddress(t *testing.T) {
	var tests = []struct {
		rawAddress string
		protocol   string
		host       string
		port       string
		err        error
	}{
		{"dns.google:853@" + Ipv6Address, "tcp-tls", Ipv6Address, "853", nil},
		{"dns.google:853@" + Ipv4Address, "tcp-tls", Ipv4Address, "853", nil},
		{Ipv4Address + ":5353", "tcp", Ipv4Address, "5353", nil},
		{Ipv6Address + ":5353", "tcp", LiteralIpa6Address, "5353", nil},
		{Ipv4Address + ":5353", "udp", Ipv4Address, "5353", nil},
		{Ipv6Address + ":5353", "udp", LiteralIpa6Address, "5353", nil},
		{Ipv4Address, "udp", Ipv4Address, "53", nil},
		{Ipv6Address, "udp", LiteralIpa6Address, "53", nil},
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
		{"socks5://" + Ipv4Address + ":80", Ipv4Address + ":80"},
		{"socks5://" + Ipv6Address + ":80", Ipv6Address + ":80"},
		{"socks5://" + Ipv6Address, Ipv6Address + ":1080"},
		{"" + Ipv4Address + ":80", Ipv4Address + ":80"},
		{"" + Ipv6Address + ":80", Ipv6Address + ":80"},
		{"" + Ipv6Address, Ipv6Address + ":1080"},
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
		{"dns.google:853@" + Ipv6Address, "dns.google", "853", Ipv6Address, nil},
		{"dns.google:853@" + Ipv4Address, "dns.google", "853", Ipv4Address, nil},
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
