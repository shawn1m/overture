# overture
[![Build Status](https://travis-ci.org/holyshawn/overture.png)](https://travis-ci.org/holyshawn/overture)

Overture is a DNS upstream switcher written in Go in order to purify DNS records.

Overture means an orchestral piece at the beginning of a classical music composition, just like DNS which is nearly the 
first step of surfing the Internet.

Overture forces IPv6 and custom domain DNS queries to use alternative DNS when applicable. Overture will first query the
 domain with listed primary DNS servers in configuration, if the answer is empty or does not match with the custom IP
 network, then overture will query the alternative DNS servers and use their answer instead.

**Please note: If you are using the binary releases, please follow the instructions in the README file with 
corresponding git version tag. The README in master branch are subject to change and does not always reflect the correct
 instructions to your binary release version.**

## Features

+ Full IPv6 support
+ IPv6 record (AAAA) redirection, especially for **CERNET IPv6** users
+ DNS upstream via TCP with custom port
+ Custom IP network filter
+ Custom domain filter, base64 decode support
+ Minimum TTL modification support
+ EDNS client subnet support
+ Static hosts support via `hosts` file
+ Cache with EDNS client subnet

## Usages

Download binary release from the [release page](https://github.com/holyshawn/overture/releases), and run:

    ./overture # Start with the default config file -> ./config.json

Or use your own config file:

    ./overture -c /xxx/xxx/x.json # Use your own config file

Verbose mode:

    ./overture -v # This will show more information
    
For other options, please see help:

    ./overture -h # This will show some parameters for help

Tips:

+ Root privilege is required if you are listening on port 53
+ You may find default IP network file and domain file from the acknowledgements section, or just download below files.
  These files are also included in the binary release package.
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
        "ExternalIP": ""
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
    + Name: This field is only used for logging
    + Protocol: `tcp` or `udp`
    + EDNSClientSubnet: Used to improve DNS accuracy. Please check [RFC7871](https://tools.ietf.org/html/rfc7871) for 
    details.
        + Policy: 
            + `auto`: If client IP is not in the reserved IP network, use client IP. Otherwise, use external IP.
            + `disable`: Disable this feature.
        + ExternalIP: If this field is empty, EDNS client subnet will be disabled when used.
+ RedirectIPv6Record: Redirect IPv6 DNS queries to alternative DNS servers.
+ File: For Windows users, you can use relative path like `.\file.txt`, or absolute path like `C:\path\to\file.txt` in
  the config.
+ DomainBase64Decode: If this file is base64 decoded, use `true`
+ MinimumTTL: Set the minimum TTL value (in seconds) in order to improve caching efficiency, use `0` to disable.

Hosts: 

+ Using wildcard `*` in the subdomain for wildcard matching is allowed, e.g. `192.168.0.2 *.db.local`.

DNS servers with EDNS client subnet support:

+ DNSPod 119.29.29.29:53
+ GoogleDNS 8.8.8.8:53 \[2001:4860:4860::8888\]:53

## Acknowledgements

+ @clowwindy: the author of the [ChinaDNS](https://github.com/shadowsocks/ChinaDNS)
+ @miekg: the author of the [dns](https://github.com/miekg/dns), and [skydns](https://github.com/skynetservices/skydns)
+ @janeczku: the author of the [go-dnsmasq](https://github.com/janeczku/go-dnsmasq)
+ @sirupsen: the author of the [logrus](https://github.com/Sirupsen/logrus)
+ @17mon: the author of the [china_ip_list](https://github.com/17mon/china_ip_list)
+ @gfwlist: the author of the [gfwlist](https://github.com/gfwlist/gfwlist)
+ Contributors: @V-E-O, @sh1r0, @maddie

## License

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text.
