package common

type DNSUpstream struct {
	Name             string
	Address          string
	Protocol         string
	SOCKS5Address    string
	Timeout          int
	EDNSClientSubnet *EDNSClientSubnetType
}
