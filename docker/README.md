# Container Support:

```
docker run -it --rm --name=overture \
  -e PRIMARY_DNS=AliDNS/dns.alidns.com:853@223.6.6.6/tcp-tls,Alidns-doh/https://dns.alidns.com/dns-query/https \
  -e ALTERNATIVE_DNS=Google/dns.google:853@8.8.4.4/tcp-tls,Cloudflare/one.one.one.one:853@1.0.0.1/tcp-tls,OpenDNS/208.67.222.222:443/tcp \
  -e CACHE_SIZE=1000 \
  -v /etc/hosts:/opt/overture/hosts_sample \
  -p 15353:53/udp sgrio/overture
```

Supported Environment Variables:
- **PRIMARY_DNS**, comma separated DNS list in format of \<name\>/\<address\>/\<protocol\>
- **ALTERNATIVE_DNS**, similar to primary DNS
- **IP_NETWORK_FILE_PRIMARY**, location of IP network file
- **DOMAIN_FILE_ALTERNATIVE**, location of domain file for alternative DNS
- **HOSTS_FILE**, location of hosts file
- **CACHE_SIZE**, cache size
- **ALTERNATIVE_DNS_SOCKS5_PROXY**, socks5 proxy for alternative DNS

You can also update config file by mounting directly into container.
