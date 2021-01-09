# overture
[![Build Status](https://travis-ci.org/shawn1m/overture.svg)](https://travis-ci.org/shawn1m/overture)
[![Build status](https://ci.appveyor.com/api/projects/status/gqrixsfcmmrcaohr/branch/master?svg=true)](https://ci.appveyor.com/project/shawn1m/overture/branch/master)
[![GoDoc](https://godoc.org/github.com/shawn1m/overture?status.svg)](https://godoc.org/github.com/shawn1m/overture)
[![Go Report Card](https://goreportcard.com/badge/github.com/shawn1m/overture)](https://goreportcard.com/report/github.com/shawn1m/overture)
[![codecov](https://codecov.io/gh/shawn1m/overture/branch/master/graph/badge.svg)](https://codecov.io/gh/shawn1m/overture)

Overture is a customized DNS relay server.

Overture means an orchestral piece at the beginning of a classical music composition, just like DNS which is nearly the
first step of surfing the Internet.

**Please note:** 
- Read the **entire README first** is necessary if you want to use overture **safely** or **create an issue** for this project .
- **Production usage is not recommended and there is no guarantee or warranty of it.**
   
## Features

+ Full IPv6 support
+ Multiple DNS upstream
    + Via UDP/TCP with custom port
    + Via SOCKS5 proxy (TCP only)
    + With EDNS Client Subnet (ECS) [RFC7871](https://tools.ietf.org/html/rfc7871)
+ Dispatcher
    + IPv6 record (AAAA) redirection
    + Custom domain
    + Custom IP network
+ Minimum TTL modification
+ Hosts (Both IPv4 and IPv6 are supported and IPs will be returned in random order. If you want to use regex match, please understand regex first)
+ Cache with ECS and Redis(Persistence) support
+ DNS over HTTP server support

### Dispatch process

DNS queries with a custom domain can be forced to use selected DNS when applicable.

For custom IP network, overture will send queries to primary DNS firstly. Then, If the answer is empty, or the IP
is not matched, the alternative DNS servers will be used instead.

## Installation

The binary releases are available in [releases](https://github.com/shawn1m/overture/releases).

## Usages

Start with the default config file `./config.yml` 

**Only file having a `.json` suffix will be considered as json format for compatibility reason, and this support is deprecated right now.**

    $ ./overture

Or use your own config file:

    $ ./overture -c /path/to/config.yml

Verbose mode:

    $ ./overture -v

Log to file:

    $ ./overture -l /path/to/overture.log

For other options, please check the helping menu:

    $ ./overture -h

Tips:

+ Root privilege might be required if you want to let overture listen on port 53 or one of other system ports.

###  Configuration Syntax

Configuration file is "config.yml" by default:

```yaml
bindAddress: :53
debugHTTPAddress: 127.0.0.1:5555
dohEnabled: false
primaryDNS:
  - name: DNSPod
    address: 119.29.29.29:53
    protocol: udp
    socks5Address:
    timeout: 6
    ednsClientSubnet:
      policy: disable
      externalIP:
      noCookie: true
alternativeDNS:
  - name: 114DNS
    address: 114.114.114.114:53
    protocol: udp
    socks5Address:
    timeout: 6
    ednsClientSubnet:
      policy: disable
      externalIP:
      noCookie: true
onlyPrimaryDNS: false
ipv6UseAlternativeDNS: false
alternativeDNSConcurrent: false
whenPrimaryDNSAnswerNoneUse: primaryDNS
ipNetworkFile:
  primary: ./ip_network_primary_sample
  alternative: ./ip_network_alternative_sample
domainFile:
  primary: ./domain_primary_sample
  alternative: ./domain_alternative_sample
  matcher: full-map
hostsFile:
  hostsFile: ./hosts_sample
  finder: full-map
minimumTTL: 0
domainTTLFile: ./domain_ttl_sample
cacheSize: 0
cacheRedisUrl: redis://localhost:6379/0
cacheRedisConnectionPoolSize: 10 
rejectQType:
  - 255
```

Tips:

+ bindAddress: Specifying only port (e.g. `:53`) will let overture listen on all available addresses (both IPv4 and
IPv6). Overture will handle both TCP and UDP requests. Literal IPv6 addresses are enclosed in square brackets (e.g. `[2001:4860:4860::8888]:53`)
+ debugHTTPAddress: Specifying an HTTP port for debugging (**`5555` is the default port despite it is also acknowledged as the android Wi-Fi adb listener port**), currently used to dump DNS cache, and the request url is `/cache`, available query argument is `nobody`(boolean)

    * true(default): only get the cache size;

        ```bash
        $ curl 127.0.0.1:5555/cache | jq
        {
          "length": 1,
          "capacity": 100,
          "body": {}
        }
        ```

    * false: get cache size along with cache detail.

        ```bash
        $ curl 127.0.0.1:5555/cache?nobody=false | jq
        {
          "length": 1,
          "capacity": 100,
          "body": {
            "www.baidu.com. 1": [
              {
                "name": "www.baidu.com.",
                "ttl": 1140,
                "type": "CNAME",
                "rdata": "www.a.shifen.com."
              },
              {
                "name": "www.a.shifen.com.",
                "ttl": 300,
                "type": "CNAME",
                "rdata": "www.wshifen.com."
              },
              {
                "name": "www.wshifen.com.",
                "ttl": 300,
                "type": "A",
                "rdata": "104.193.88.123"
              },
              {
                "name": "www.wshifen.com.",
                "ttl": 300,
                "type": "A",
                "rdata": "104.193.88.77"
              }
            ]
          }
        }
        ```
+ dohEnabled: Enable DNS over HTTP server using `DebugHTTPAddress` above with url path `/dns-query`. DNS over HTTPS server can be easily achieved helping by another web server software like caddy or nginx. (Experimental)
+ primaryDNS/alternativeDNS:
    + name: This field is only used for logging.
    + address: Same rule as BindAddress.
    + protocol: `tcp`, `udp`, `tcp-tls` or `https`
        + `tcp-tls`: Address format is "servername:port@serverAddress", try one.one.one.one:853 or one.one.one.one:853@1.1.1.1
        + `https`: Just try https://cloudflare-dns.com/dns-query
        +  Check [DNS Privacy Public Resolvers](https://dnsprivacy.org/wiki/display/DP/DNS+Privacy+Public+Resolvers) for more public `tcp-tls`, `https` resolvers.
    + socks5Address: Forward dns query to this SOCKS5 proxy, `“”` to disable.
    + ednsClientSubnet: Use this to improve DNS accuracy for many reasons. Please check [RFC7871](https://tools.ietf.org/html/rfc7871) for
    details.
        + policy
            + `auto`: If the client IP is not in the reserved IP network, use the client IP. Otherwise, use the external IP.
            + `manual`: Use the external IP if this field is not empty, otherwise use the client IP if it is not one of the reserved IPs.
            + `disable`: Disable this feature.
        + externalIP: If this field is empty, ECS will be disabled when the inbound IP is not an external IP.
        + noCookie: Disable cookie.
+ onlyPrimaryDNS: Disable dispatcher feature, use primary DNS only.
+ ipv6UseAlternativeDNS: For to redirect IPv6 DNS queries to alternative DNS servers.
+ alternativeDNSConcurrent: Query the primaryDNS and alternativeDNS at the same time.
+ whenPrimaryDNSAnswerNoneUse: If the response of primaryDNS exists and there is no `ANSWER SECTION` in it, the final chosen DNS upstream should be defined here. (There is no `AAAA` record for most domains right now) 
+ *File: Both relative like `./file` or absolute path like `/path/to/file` are supported. Especially, for Windows users, please use properly escaped path like
  `C:\\path\\to\\file.txt` in the configuration.
+ domainFile.Matcher: Matching policy and implementation, including "full-list", "full-map", "regex-list", "mix-list", "suffix-tree" and "final". Default value is "full-map".
+ hostsFile.Finder: Finder policy and implementation, including "full-map", "regex-list". Default value is "full-map".
+ domainTTLFile: Regex match only for now;
+ minimumTTL: Set the minimum TTL value (in seconds) in order to improve caching efficiency, use `0` to disable.
+ cacheSize: The number of query record to cache, use `0` to disable.
+ cacheRedisUrl, cacheRedisConnectionPoolSize: Use redis cache instead of local cache. (Experimental)
+ rejectQType: Reject query with specific DNS record types, check [List of DNS record types](https://en.wikipedia.org/wiki/List_of_DNS_record_types) for details.

#### Domain file example (full match)

    example.com

#### Domain file example (regex match)

    ^xxx.xx
    
#### IP network file example (CIDR match)

    1.0.1.0/24
    ::1/128
    
#### Domain TTL file example (regex match)
 
     example.com$ 100

#### Hosts file example (full match)

    127.0.0.1 localhost
    ::1 localhost
    
#### Hosts file example (regex match)

    10.8.0.1 example.com$

#### DNS servers with ECS support

+ DNSPod 119.29.29.29:53

For DNSPod, ECS might only work via udp, you can test it by [patched dig](https://www.gsic.uva.es/~jnisigl/dig-edns-client-subnet.html) to certify this argument by comparing answers.
 
**The accuracy depends on the server side.**

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
;www.qq.com.            IN  A

;; ANSWER SECTION:
www.qq.com.     300 IN  A   101.226.103.106

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
;www.qq.com.            IN  A

;; ANSWER SECTION:
www.qq.com.     43  IN  A   59.37.96.63
www.qq.com.     43  IN  A   14.17.32.211
www.qq.com.     43  IN  A   14.17.42.40

;; Query time: 81 msec
;; SERVER: 119.29.29.29#53(119.29.29.29)
;; WHEN: Wed Mar 08 18:01:32 CST 2017
;; MSG SIZE  rcvd: 87
```

## Acknowledgements

+ [dns](https://github.com/miekg/dns): BSD-3-Clause
+ [skydns](https://github.com/skynetservices/skydns): MIT
+ [go-dnsmasq](https://github.com/janeczku/go-dnsmasq):  MIT
+ [All Contributors](https://github.com/shawn1m/overture/graphs/contributors)

## License

This project is under the MIT license. See the [LICENSE](LICENSE) file for the full license text.
