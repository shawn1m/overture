# overture
[![Build Status](https://travis-ci.org/holyshawn/overture.png)](https://travis-ci.org/holyshawn/overture)

Overture is a dns upstream switcher written in golang in order to purify dns records.

Overture means an orchestral piece at the beginning of a classical music composition, just like dns which is nearly the first step of surfing the internet.

Overture forces IPv6 and custom domain DNS queries to use alternative dns. Normally, when overture is using primary dns, if response answer is empty or is not matched with custom ip network, then overture will use alternative dns instead.

**Warn: If you are using the release version, just try to follow the README file from compatible version branch tag, this README file is always in development.**

## Features

+ Full IPv6 support
+ IPv6 record (AAAA) redirection, especially for **CERNET IPv6** users
+ TCP dns upstream with custom port
+ Custom IP network filter
+ Custom domain filter, base64 decode support
+ Minimum TTL modification support
+ EDNS client subnet support
+ Hosts support
+ Cache with EDNS client subnet

## Usages

Download from the [release](https://github.com/holyshawn/overture/releases), just run:

    ./overture # Start with the default config file -> ./config.json

Or use your own config file:

    ./overture -c /xxx/xxx/x.json # Use your own config file

Verbose mode:

    ./overture -v # This will show more information
    
Extra Help:

    ./overture -h # This will show some parameters for help

Tips:

+ You may need sudo to start overture on port 53.
+ You may find default IP network file and domain file from acknowledgements or just download below files. These files are also included in the release.
+ [ip_network file ](https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt)
+ [base64 domain file](https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt)

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
      "Timeout": 6,
      "EDNSClientSubnet": {
        "Policy": "disable",
        "ExternalIP": "",
        "CustomIP": ""
      }
    }
  ],
  "AlternativeDNS":[
    {
      "Name": "OpenDNS",
      "Address": "208.67.222.222:443",
      "Protocol": "tcp",
      "Timeout": 6,
      "EDNSClientSubnet":{
        "Policy": "disable",
        "ExternalIP": "",
        "CustomIP": ""
      }
    }
  ],
  "RedirectIPv6Record": true,
  "IPNetworkFile": "/xx/xx.txt",
  "DomainFile": "/xx/xx.txt",
  "DomainBase64Decode": true,
  "HostsFile": "./hosts",
  "MinimumTTL": 0,
  "CacheSize" : 0
}
```

Tips:

+ BindAddress: No IP means listen both IPv4 and IPv6, overture will listen both TCP and UDP ports.
+ DNS: You Can use multiple dns upstream in this list.
    + Name: Just for log.
    + Protocol: "tcp" or "udp".
    + EDNSClientSubnet: Improve DNS accuracy. [RFC7871](https://tools.ietf.org/html/rfc7871)
        + Policy: 
            + auto: If client IP is not in the reserved ip network, use client IP. Otherwise, use external IP.
            + custom: Always use custom ip.
            + disable: Disable this feature.
        + ExternalIP: If this field is empty, edns client subnet will be disabled when use it.
+ RedirectIPv6Record: Redirect IPv6 DNS query to alternative dns.
+ Path: For windows user, if you want to use absolute path, please use like this: "C:\\\xx\\\xx.txt".
+ DomainBase64Decode: Could be empty field.
+ MinimumTTL: Set the minimum TTL value (second) in order to improve cache, use 0 to disable.

Hosts: 

A wildcard * in the left-most label of hostnames is allowed, like 192.168.0.2 *.db.local.

DNS with EDNS client subnet:

+ DNSPod 119.29.29.29:53
+ OpenDNS 208.67.222.222:443 \[2620:0:ccc::2\]:443
+ GoogleDNS 8.8.8.8:53 \[2001:4860:4860::8888\]:53

## Acknowledgements

+ @clowwindy: the author of the [ChinaDNS](https://github.com/shadowsocks/ChinaDNS)
+ @miekg: the author of the [dns](https://github.com/miekg/dns), and [skydns](https://github.com/skynetservices/skydns)
+ @janeczku: the author of the [go-dnsmasq](https://github.com/janeczku/go-dnsmasq)
+ @sirupsen: the author of the [logrus](https://github.com/Sirupsen/logrus)
+ @17mon: the author of the [china_ip_list](https://github.com/17mon/china_ip_list)
+ @gfwlist: the author of the [gfwlist](https://github.com/gfwlist/gfwlist)
+ Pull requests: @V-E-O, @sh1r0

## License

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text.
