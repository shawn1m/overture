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

// support two formats: scheme://127.0.0.1:1080 or 127.0.0.1:1080
func extractUrl(rawAddress string, protocol string) (host string, port string, err error) {

	if !strings.Contains(rawAddress, "://") {
		rawAddress = protocol + "://" + rawAddress
	}

	uri, err := url.Parse(rawAddress)
	if err != nil {
		log.Warnf("url %s is invalid", rawAddress)
		return "", "", errors.New("url is invalid")
	}
	host = uri.Hostname()

	if len(uri.Scheme) == 0 || uri.Scheme != protocol {
		return "", "", errors.New("url is invalid")
	}

	port = uri.Port()
	if len(port) == 0 {
		port = getDefaultPort(protocol)
	}
	return
}

func ExtractFullUrl(rawAddress string, protocol string) (string, error) {
	host, port, err := extractUrl(rawAddress, protocol)
	return net.JoinHostPort(host, port), err
}

func extractTLSDNSAddress(rawAddress string, protocol string) (host string, port string, err error) {
	rawAddress = protocol + "://" + rawAddress
	s := strings.Split(rawAddress, "@")

	host, port, err = extractUrl(s[0], protocol)

	if err != nil {
		return "", "", nil
	}

	if len(s) == 2 && isJustIP(s[1]) {
		host = generateLiteralIPv6AddressIfNecessary(s[1])
	} else {
		log.Warnf("dns server address %s is invalid", rawAddress)
		return "", "", errors.New("dns up server address is invalid")
	}
	return host, port, nil
}

func ExtractTLSDNSHostName(rawAddress string) (host string, err error) {
	rawAddress = "tcp-tls" + "://" + rawAddress
	s := strings.Split(rawAddress, "@")

	host, _, err = extractUrl(s[0], "tcp-tls")
	return host, err
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

// ExtractDNSAddress parse all format, return literal IPv6 address
func ExtractDNSAddress(rawAddress string, protocol string) (host string, port string, err error) {
	switch protocol {
	case "tcp-tls":
		host, port, err = extractTLSDNSAddress(rawAddress, protocol)
	default:
		host, port, err = extractUrl(rawAddress, protocol)
	}
	return host, port, err
}
