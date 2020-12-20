package common

type DNSUpstream struct {
	Name             string                `yaml:"name"`
	Address          string                `yaml:"address"`
	Protocol         string                `yaml:"protocol"`
	SOCKS5Address    string                `yaml:"socks5Address"`
	Timeout          int                   `yaml:"timeout"`
	EDNSClientSubnet *EDNSClientSubnetType `yaml:"ednsClientSubnet"`
	TCPPoolConfig    struct {
		Enable          bool `yaml:"enable"`
		InitialCapacity int  `yaml:"initialCapacity"`
		MaxCapacity     int  `yaml:"maxCapacity"`
		IdleTimeout     int  `yaml:"idleTimeout"`
	} `yaml:"tcpPoolConfig"`
}
