# overture
[![Build Status](https://travis-ci.org/holyshawn/overture.png)](https://travis-ci.org/holyshawn/overture)

Overture is a lightweight upstream dns switcher written in golang in order to purify dns records.

Overture means an orchestral piece at the beginning of a classical music composition, just like dns which is nearly the first step of surfing the internet.

Overture force IPv6 and custom domain DNS queries to use alternative dns. Normally, when overture is using primary dns, if response answer is empty or is not matched with custom ip network then use alternative dns.

**Warn: If you use the release version, just try to follow the README file from compatible version tags, this README file is usually in development.**

## Features

+ Full IPv6 support
+ IPv6 record redirection, especially for **CERNET IPv6** users
+ TCP upstream dns server with custom port
+ Custom IP network filter
+ Custom domain filter, base64 decode support
+ Minimum TTL setting support
+ EDNS client subnet support

## Usages

Download from the [release](https://github.com/holyshawn/overture/releases), just run:

    ./overture # Start with the default config file -> ./config.json

Or use your own config file:

    ./overture -c /xxx/xxx/x.json # Use your own config file

Verbose mode :

    ./overture -v # This will show more information
    
Extra Help:

    ./overture -h # This will show some parameters for help

Tips:

+ You may need sudo to start overture on port 53.
+ You may find default IP network file and domain file from acknowledgements or just download below files. These files are also included in the release.
+ [ip_network file ](https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt)
+ [base64 domain file](https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt)
+ You may need some third-party software such as dnsmasq to cache your dns records.

###  Configuration Syntax

Configuration file is "config.json":

```json
{
  "BindAddress": ":53",
  "PrimaryDNSAddress": "119.29.29.29:53",
  "PrimaryDNSProtocol": "udp",
  "AlternativeDNSAddress": "208.67.222.222:443",
  "AlternativeDNSProtocol": "tcp",
  "Timeout": 6,
  "RedirectIPv6Record": true,
  "IPNetworkFilePath": "/xx/xx.txt",
  "DomainFilePath": "/xx/xx.txt",
  "DomainBase64Decode": true,
  "MinimumTTL": -1,
  "EDNSClientSubnetPolicy": "disable",
  "EDNSClientSubnetIP": ""
}
```

Tips:

+ BindAddress: No IP means listen both IPv4 and IPv6, overture will listen both TCP and UDP ports
+ DNS:
    + DNSPod 119.29.29.29:53
    + OpenDNS 208.67.222.222:443 \[2620:0:ccc::2\]:443
+ Protocol: "tcp" or "udp"
+ RedirectIPv6Record: Redirect IPv6 DNS query to alternative dns
+ Path: For windows user, please use path like "C:\\xx\\xx.txt"
+ DomainBase64Decode: Could be empty field
+ MinimumTTL: Set the minimum TTL value (second) in order to improve cache, use -1 to disable.
+ EDNSClientSubnetPolicy: Improve DNS accuracy, only works for primary dns. [RFC7871](https://tools.ietf.org/html/rfc7871)
    + auto: If client IP is not in the reserved ip network, use client IP. Otherwise, use server external IP.
    + custom: Always use EDNSClientSubnetIP.
    + disable: Disable this feature.
    + DNSPod, OpenDNS and GoogleDNS support this feature.

## To Do

+ ~~edns support~~
+ ~~ttl revision support~~

## Acknowledgements

+ @clowwindy: the author of the [ChinaDNS](https://github.com/shadowsocks/ChinaDNS)
+ @miekg: the author of the [dns](https://github.com/miekg/dns)
+ @sirupsen: the author of the [logrus](https://github.com/Sirupsen/logrus)
+ @17mon: the author of the [china_ip_list](https://github.com/17mon/china_ip_list)
+ @gfwlist: the author of the [gfwlist](https://github.com/gfwlist/gfwlist)

## License

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text
