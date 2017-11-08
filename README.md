# overture
[![Build Status](https://travis-ci.org/shawn1m/overture.svg)](https://travis-ci.org/shawn1m/overture)
[![GoDoc](https://godoc.org/github.com/shawn1m/overture?status.svg)](https://godoc.org/github.com/shawn1m/overture)
[![Go Report Card](https://goreportcard.com/badge/github.com/shawn1m/overture)](https://goreportcard.com/report/github.com/shawn1m/overture)

Overture is a DNS server/forwarder/dispatcher written in Go.

Overture means an orchestral piece at the beginning of a classical music composition, just like DNS which is nearly the
first step of surfing the Internet.

**Please note: If you are using the binary releases, please follow the instructions in the README file with
corresponding git version tag. The README in master branch are subject to change and does not always reflect the correct
 instructions to your binary release version.**

## Features

+ Full IPv6 support
+ Multiple DNS upstream
    + Via UDP/TCP with custom port
    + Via SOCKS5 proxy (TCP only)
    + With EDNS Client Subnet (ECS) [RFC7871](https://tools.ietf.org/html/rfc7871)
+ Dispatcher
    + IPv6 record (AAAA) redirection
    + Custom IP network
    + Custom domain
+ Minimum TTL modification
+ Hosts (prefix wildcard, random order of multiple answers)
+ Cache with ECS

### Dispatch process

Overture forces IPv6 and custom domain DNS queries to use alternative DNS when applicable.

As for custom IP network, overture will first query the domain with primary DNS, if the answer is empty or the IP
is not matched then overture will query the alternative DNS servers and use their answer instead.

## Installation

You can download binary releases from the [release](https://github.com/shawn1m/overture/releases).

For ArchLinux users, package `overture` is available in AUR. If you use a AUR helper i.e. `yaourt`, you can simply run:

    yaourt -S overture

For mips users, please assure the kernel FPU emulation is enabled, check [#32](https://github.com/shawn1m/overture/issues/32) [#26](https://github.com/shawn1m/overture/issues/26) [golang/go#18880](https://github.com/golang/go/issues/18880) for details.

## Usages

Start with the default config file -> ./config.json

    $ ./overture

Or use your own config file:

    $ ./overture -c /path/to/config.json

Verbose mode:

    $ ./overture -v

Log to file:

    $ ./overture -l /path/to/overture.log

For other options, please see help:

    $ ./overture -h

Tips:

+ Root privilege is required if you are listening on port 53.
+ For Windows users, you can run overture on command prompt instead of double click.

###  Configuration Syntax

Configuration file is "config.json" by default:

```json
{
  "BindAddress": ":53",
  "PrimaryDNS": [
    {
      "Name": "DNSPod",
      "Address": "119.29.29.29:53",
      "Protocol": "udp",
      "SOCKS5Address": "",
      "Timeout": 6,
      "EDNSClientSubnet": {
        "Policy": "disable",
        "ExternalIP": ""
      }
    }
  ],
  "AlternativeDNS": [
    {
      "Name": "OpenDNS",
      "Address": "208.67.222.222:443",
      "Protocol": "tcp",
      "SOCKS5Address": "",
      "Timeout": 6,
      "EDNSClientSubnet": {
        "Policy": "disable",
        "ExternalIP": ""
      }
    }
  ],
  "OnlyPrimaryDNS": false,
  "RedirectIPv6Record": false,
  "IPNetworkFile": "./ip_network_sample",
  "DomainFile": "./domain_sample",
  "DomainBase64Decode": true,
  "HostsFile": "./hosts_sample",
  "MinimumTTL": 0,
  "CacheSize" : 0,
  "RejectQtype": [255]
}
```

Tips:

+ BindAddress: Specifying only port (e.g. `:53`) will have overture listen on all available addresses (both IPv4 and
IPv6). Overture will handle both TCP and UDP requests. Literal IPv6 addresses are enclosed in square brackets (e.g. `[2001:4860:4860::8888]:53`)
+ DNS: You can specify multiple DNS upstream servers here.
    + Name: This field is only used for logging.
    + Address: Same as BindAddress.
    + Protocol: `tcp` or `udp`
    + SOCKS5Address: Forward dns query to this SOCKS5 proxy, `“”` to disable.
    + EDNSClientSubnet: Used to improve DNS accuracy. Please check [RFC7871](https://tools.ietf.org/html/rfc7871) for
    details.
        + Policy
            + `auto`: If client IP is not in the reserved IP network, use client IP. Otherwise, use external IP.
            + `disable`: Disable this feature.
        + ExternalIP: If this field is empty, ECS will be disabled when the inbound IP is not an external IP.
+ OnlyPrimaryDNS: Disable dispatcher feature, use primary DNS only.
+ RedirectIPv6Record: Redirect IPv6 DNS queries to alternative DNS servers.
+ File: Absolute path like `/path/to/file` is allowed. For Windows users, please use properly escaped path like
  `C:\\path\\to\\file.txt` in the configuration.
+ MinimumTTL: Set the minimum TTL value (in seconds) in order to improve caching efficiency, use `0` to disable.
+ CacheSize: The number of query record to cache, use `0` to disable.
+ RejectQtype: Reject inbound query with specific DNS record types, check [List of DNS record types](https://en.wikipedia.org/wiki/List_of_DNS_record_types) for details.

#### Domain file example (Find domains and suffix match)

    example.com
    xxx.xx

#### IP network file example

    1.0.1.0/24
    10.8.0.0/16
    ::1/128

#### Hosts file example (Support prefix wildcard only, *.xxx.xx includes xxx.xx)

    127.0.0.1 localhost
    ::1 localhost
    10.8.0.1 example.com
    192.168.0.2 *.xxx.xx

#### DNS servers with ECS support

+ DNSPod 119.29.29.29:53

**For DNSPod, ECS only works via udp, you can test it by [patched dig](https://www.gsic.uva.es/~jnisigl/dig-edns-client-subnet.html)**

You can compare the response IP with the client IP to test the feature. The accuracy depends on the server side.

```
$ dig @119.29.29.29 www.qq.com +client=119.29.29.29

; <<>> DiG 9.9.3 <<>> @119.29.29.29 www.qq.com +client=119.29.29.29
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 64995
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
; CLIENT-SUBNET: 119.29.29.29/32/24
;; QUESTION SECTION:
;www.qq.com.			IN	A

;; ANSWER SECTION:
www.qq.com.		300	IN	A	101.226.103.106

;; Query time: 52 msec
;; SERVER: 119.29.29.29#53(119.29.29.29)
;; WHEN: Wed Mar 08 18:00:52 CST 2017
;; MSG SIZE  rcvd: 67
```

```
$ dig @119.29.29.29 www.qq.com +client=119.29.29.29 +tcp

; <<>> DiG 9.9.3 <<>> @119.29.29.29 www.qq.com +client=119.29.29.29 +tcp
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 58331
;; flags: qr rd ra; QUERY: 1, ANSWER: 3, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;www.qq.com.			IN	A

;; ANSWER SECTION:
www.qq.com.		43	IN	A	59.37.96.63
www.qq.com.		43	IN	A	14.17.32.211
www.qq.com.		43	IN	A	14.17.42.40

;; Query time: 81 msec
;; SERVER: 119.29.29.29#53(119.29.29.29)
;; WHEN: Wed Mar 08 18:01:32 CST 2017
;; MSG SIZE  rcvd: 87
```

## Acknowledgements

+ Dependencies:
    + [dns](https://github.com/miekg/dns): BSD-3-Clause
    + [logrus](https://github.com/Sirupsen/logrus): MIT
+ Code reference:
    + [skydns](https://github.com/skynetservices/skydns): MIT
    + [go-dnsmasq](https://github.com/janeczku/go-dnsmasq):  MIT
+ Contributors: @V-E-O, @sh1r0, @maddie, @hexchain, @everfly, @simonsmh

## License

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text.
