# overture
[![Build Status](https://travis-ci.org/holyshawn/overture.png)](https://travis-ci.org/holyshawn/overture)

Overture is a light weight upstream dns switcher written in golang in order to purify dns records.

Overture means an orchestral piece at the beginning of an classical music composition, just like dns which is nearly the first step of surfing the internet.

Overture force IPv6 DNS question and custom domain to use alternative dns, if response answer is matched with custom ip network, use primary dns otherwise use alternative dns.

## Features

+ Full IPv6 support
+ IPv6 record redirection, especially for **CERNET IPv6** users
+ TCP upstream dns server with custom port
+ Custom IP network filter
+ Custom domain filter, base64 decode support

## Usages

Download from the [release](https://github.com/holyshawn/overture/releases), just run:

    ./overture # Start with the default config file -> ./config.json

Or use your own config directory:

    ./overture -c /xxx/xxx/x.json # Use your own config file

Verbose mode :

    ./overture -v # This will show more information
    
Extra Help:

    ./overture -h # This will show some parameters for help

Tips:

+ You may need sudo to start overture on port 53.
+ You may find default IP network  file and domain file from acknowledgements parts or just download here.
+ [ip_network file ](https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt)
+ [base64 domain file](https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt)
+ You may need some third-party software such as dnsmasq to cache your dns records.

###  Configuration Syntax

Configuration file is "config.json":

```json
{
  "BindAddress": ":53",
  "PrimaryDNSAddress": "114.114.114.114:53",
  "PrimaryDNSMethod": "udp",
  "AlternativeDNSAddress": "208.67.222.222:443",
  "AlternativeDNSMethod": "tcp",
  "Timeout": 6,
  "RedirectIPv6Record": true,
  "IPNetworkFilePath": "/xx/xx.txt",
  "DomainFilePath": "/xx/xx.txt",
  "DomainBase64Decode": true
}
```

Tips:

+ BindAddress: No IP means listen both IPv4 and IPv6
+ DNS:
    + 114DNS 114.114.114.114:53
    + OpenDNS 208.67.222.222:443 \[2620:0:ccc::2\]:443
+ RedirectIPv6Record: Redirect IPv6 DNS Question to alternative dns
+ Path: For windows user, please use path like "C:\\xx\\xx.txt"
+ DomainBase64Decode: Could be empty field

## To Do

+ edns support
+ ttl revision support

## Acknowledgements

+ @clowwindy: the author of the [ChinaDNS](https://github.com/shadowsocks/ChinaDNS)
+ @miekg: the author of the [dns](https://github.com/miekg/dns)
+ @sirupsen: the author of the [logrus](https://github.com/Sirupsen/logrus)
+ @17mon: the author of the [china_ip_list](https://github.com/17mon/china_ip_list)
+ @gfwlist: the author of the [gfwlist](https://github.com/gfwlist/gfwlist)

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text