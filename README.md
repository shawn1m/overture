# overture
[![Build Status](https://travis-ci.org/shawn1m/overture.svg)](https://travis-ci.org/shawn1m/overture)

Overture is a DNS dispatcher written in Go.

Overture means an orchestral piece at the beginning of a classical music composition, just like DNS which is nearly the
first step of surfing the Internet.

**Please note: If you are using the binary releases, please follow the instructions in the README file with
corresponding git version tag. The README in master branch are subject to change and does not always reflect the correct
 instructions to your binary release version.**

## Features

+ Full IPv6 support
+ Multiple DNS upstream:
    + Via UDP/TCP with custom port
    + Via SOCKS5 proxy
    + With EDNS client subnet [RFC7871](https://tools.ietf.org/html/rfc7871)
+ Dispatcher:
    + IPv6 record (AAAA) redirection
    + Custom IP network
    + Custom domain with base64 decode support
+ Minimum TTL modification
+ Static hosts support via `hosts` file
+ Cache with EDNS client subnet support

#### Dispatch process

Overture forces IPv6 and custom domain DNS queries to use alternative DNS when applicable.

As for custom IP network, overture will first query the domain with primary DNS, if the answer is empty or the IP
is not matched then overture will query the alternative DNS servers and use their answer instead.

## Installation

+ You can download binary releases from the [release](https://github.com/shawn1m/overture/releases).
+ For ArchLinux users, package `overture` is available in AUR. If you use a AUR helper i.e. `yaourt`, you can simply run:

        yaourt -S overture

## Usages

Start with the default config file -> ./config.json

    ./overture

Or use your own config file:

    ./overture -c /path/to/config_file

Verbose mode:

    ./overture -v

For other options, please see help:

    ./overture -h

Tips:

+ Root privilege is required if you are listening on port 53.
+ For Windows users, you can run overture on command prompt instead of double click.
+ You can download sample IP network file and domain file from below.
  These files are also included in the binary release package.
  + [ip_network_file ](https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt)
  + [base64_domain_file](https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt)

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
  "AlternativeDNS":[
    {
      "Name": "OpenDNS",
      "Address": "208.67.222.222:443",
      "Protocol": "tcp",
      "SOCKS5Address": "",
      "Timeout": 6,
      "EDNSClientSubnet":{
        "Policy": "disable",
        "ExternalIP": ""
      }
    }
  ],
  "RedirectIPv6Record": true,
  "IPNetworkFile": "/path/to/ip_network_file",
  "DomainFile": "/path/to/domain_file",
  "DomainBase64Decode": true,
  "HostsFile": "/path/to/hosts_file",
  "MinimumTTL": 0,
  "CacheSize" : 0
}
```

Tips:

+ BindAddress: Specifying only port (e.g. `:53`) will have overture listen on all available addresses (both IPv4 and
IPv6). Overture will handle both TCP and UDP requests.
+ DNS: You can specify multiple DNS upstream servers here.
    + Name: This field is only used for logging.
    + Protocol: `tcp` or `udp`
    + SOCKS5Address: Forward dns query to this socks5 proxy, `“”` to disable.
    + EDNSClientSubnet: Used to improve DNS accuracy. Please check [RFC7871](https://tools.ietf.org/html/rfc7871) for
    details.
        + Policy:
            + `auto`: If client IP is not in the reserved IP network, use client IP. Otherwise, use external IP.
            + `disable`: Disable this feature.
        + ExternalIP: If this field is empty, EDNS client subnet will be disabled when used.
+ RedirectIPv6Record: Redirect IPv6 DNS queries to alternative DNS servers.
+ File: For Windows users, you can use relative path like `./file.txt`, or properly escaped absolute path like
  `C:\\path\\to\\file.txt` in the config.
+ DomainBase64Decode: If this file is base64 decoded, use `true`.
+ MinimumTTL: Set the minimum TTL value (in seconds) in order to improve caching efficiency, use `0` to disable.

#### Domain file example (suffix match)

    abc.com
    example.net

#### IP network file example

    1.0.1.0/24
    1.0.2.0/23

#### Hosts file example

    10.8.0.1 example.com
    192.168.0.2 *.db.local

#### DNS servers with EDNS client subnet support

+ DNSPod 119.29.29.29:53
+ GoogleDNS 8.8.8.8:53 \[2001:4860:4860::8888\]:53

## Acknowledgements

+ Dependencies:
    + [dns](https://github.com/miekg/dns): BSD-3-Clause
    + [logrus](https://github.com/Sirupsen/logrus): MIT
+ Code reference:
    + [skydns](https://github.com/skynetservices/skydns): MIT
    + [go-dnsmasq](https://github.com/janeczku/go-dnsmasq):  MIT
+ Contributors: @V-E-O, @sh1r0, @maddie, @hexchain, @everfly

## License

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text.
