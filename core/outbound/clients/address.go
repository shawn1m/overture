/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

// Package outbound implements multiple dns client and dispatcher for outbound connection.
package clients

import (
	"errors"
	"net"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

func getDefaultPort(protocol string) (port string) {
	switch protocol {
	case "udp", "tcp":
		port = "53"
	case "tcp-tls":
		port = "853"
	case "https":
		port = "443"
	}
	return port
}

// ToNetwork convert dns protocol to network
func ToNetwork(protocol string) string {
	switch protocol {
	case "udp":
		return "udp"
	case "tcp", "tcp-tls", "https":
		return "tcp"
	default:
		return ""
	}
}

// ExtractSocksAddress parse socks5 address,
// suport two formats: socks5://127.0.0.1:1080 and 127.0.0.1:1080
func ExtractSocksAddress(rawAddress string) (string, error) {
	uri, err := url.Parse(rawAddress)
	if err != nil {
		// socks5 address format is 127.0.0.1:1080
		s := strings.Split(rawAddress, ":")
		if len(s) == 2 {
			return net.JoinHostPort(s[0], s[1]), nil
		}
		return net.JoinHostPort(s[0], "1080"), nil
	}
	// socks5://127.0.0.1:1080
	if len(uri.Scheme) == 0 || uri.Scheme != "socks5" {
		return "", errors.New("socks5 address is invalid")
	}
	port := uri.Port()
	if len(port) == 0 {
		port = "1080"
	}
	address := net.JoinHostPort(uri.Hostname(), port)
	return address, nil
}

// ExtractTLSDNSAddress parse tcp-tls format: dns.google:853@8.8.8.8
func ExtractTLSDNSAddress(rawAddress string) (host string, port string, ip string) {
	s := strings.Split(rawAddress, "@")
	if len(s) == 2 {
		ip = s[1]
	}
	items := strings.Split(s[0], ":")
	host = items[0]
	if len(items) == 1 {
		port = getDefaultPort("tcp-tls")
	} else {
		port = items[1]
	}
	return host, port, ip
}

// ExtractNormalDNSAddress parse normal format: 8.8.8.8:53
func ExtractNormalDNSAddress(rawAddress string, protocol string) (host string, port string, err error) {
	s := strings.Split(rawAddress, ":")
	if len(s) != 1 && len(s) != 2 {
		log.Warnf("dns server address %s is invalid", rawAddress)
		return "", "", errors.New("dns up server adrress is invalid")
	}
	host = s[0]
	if len(s) == 2 {
		port = s[1]
	} else {
		port = getDefaultPort(protocol)
	}
	return host, port, nil

}

// ExtractHTTPSAddress parse https format: https://dns.google/dns-query
func ExtractHTTPSAddress(rawAddress string) (host string, port string, err error) {
	uri, err := url.Parse(rawAddress)
	if err != nil {
		return "", "", err
	}
	host = uri.Hostname()
	port = uri.Port()
	if len(port) == 0 {
		port = getDefaultPort("https")
	}
	return host, port, nil

}

// ExtractDNSAddress parse all format
func ExtractDNSAddress(rawAddress string, protocol string) (host string, port string, err error) {
	switch protocol {
	case "https":
		host, port, err = ExtractHTTPSAddress(rawAddress)
	case "tcp-tls":
		host, port, _ = ExtractTLSDNSAddress(rawAddress)
	default:
		host, port, err = ExtractNormalDNSAddress(rawAddress, protocol)
	}
	return host, port, err
}
