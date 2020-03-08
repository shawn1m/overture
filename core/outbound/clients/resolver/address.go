/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

// Package outbound implements multiple dns client and dispatcher for outbound connection.
package resolver

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
	case "socks5":
		port = "1080"
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
// support two formats: socks5://127.0.0.1:1080 and 127.0.0.1:1080
func ExtractSocksAddress(rawAddress string) (string, error) {
	uri, err := url.Parse(rawAddress)
	if err != nil {
		// socks5 address format is 127.0.0.1:1080
		_, _, err = net.SplitHostPort(rawAddress)
		isJustIP := isJustIP(rawAddress)
		if err != nil && !isJustIP {
			log.Warnf("socks5 address %s is invalid", rawAddress)
			return "", errors.New("socks5 address is invalid")
		}
		if isJustIP {
			rawAddress = rawAddress + ":" + getDefaultPort("socks5")
		}
		return rawAddress, nil
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
func ExtractTLSDNSAddress(rawAddress string) (host string, port string, ip string, err error) {
	s := strings.Split(rawAddress, "@")
	host, port, err = net.SplitHostPort(s[0])
	isJustHost := len(rawAddress) > 0
	if err != nil && !isJustHost {
		log.Warnf("dns server address %s is invalid", rawAddress)
		return "", "", "", errors.New("dns up server address is invalid")
	}
	if err != nil && isJustHost {
		host = s[0]
		if isJustIP(host) {
			host = generateLiteralIPv6AddressIfNecessary(host)
		}
		port = getDefaultPort("tcp-tls")
	}

	ip = s[1]
	if isJustIP(ip) {
		ip = generateLiteralIPv6AddressIfNecessary(ip)
	} else {
		log.Warnf("dns server address %s is invalid", rawAddress)
		return "", "", "", errors.New("dns up server address is invalid")
	}
	return host, port, ip, nil
}

// extractNormalDNSAddress parse normal format: 8.8.8.8:53
func extractNormalDNSAddress(rawAddress string, protocol string) (host string, port string, err error) {
	host, port, err = net.SplitHostPort(rawAddress)
	isJustIP := isJustIP(rawAddress)
	if err != nil && !isJustIP {
		log.Warnf("dns server address %s is invalid", rawAddress)
		return "", "", errors.New("dns up server address is invalid")
	}
	if isJustIP {
		host = generateLiteralIPv6AddressIfNecessary(rawAddress)
		port = getDefaultPort(protocol)
	}
	return host, port, nil

}

func isJustIP(rawAddress string) bool {
	// If this rawAddress is not like "[::1]:5353", change [::1] to ::1
	if !strings.Contains(rawAddress, "]:") {
		rawAddress = generateLiteralIPv6AddressIfNecessary(rawAddress)
	}
	return net.ParseIP(rawAddress) != nil
}

func generateLiteralIPv6AddressIfNecessary(rawAddress string) string {
	rawAddress = strings.Replace(rawAddress, "[", "", 1)
	rawAddress = strings.Replace(rawAddress, "]", "", 1)
	return rawAddress
}

// extractHTTPSAddress parse https format: https://dns.google/dns-query
func extractHTTPSAddress(rawAddress string) (host string, port string, err error) {
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

// ExtractDNSAddress parse all format, return literal IPv6 address
func ExtractDNSAddress(rawAddress string, protocol string) (host string, port string, err error) {
	switch protocol {
	case "https":
		host, port, err = extractHTTPSAddress(rawAddress)
	case "tcp-tls":
		_host, _port, _ip, _err := ExtractTLSDNSAddress(rawAddress)
		if len(_ip) > 0 {
			host = _ip
		} else {
			host = _host
		}
		port = _port
		err = _err
	default:
		host, port, err = extractNormalDNSAddress(rawAddress, protocol)
	}
	return host, port, err
}
