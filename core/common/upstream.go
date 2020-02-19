package common

type DNSUpstream struct {
	Name             string
	Address          string
	Protocol         string
	SOCKS5Address    string
	Timeout          int
	EDNSClientSubnet *EDNSClientSubnetType
	TCPPoolConfig    struct {
		Enable          bool
		InitialCapacity int
		MaxCapacity     int
		IdleTimeout     int
	}
}
