#!/bin/sh
set -e

replacePrimaryDns() {
  yq eval 'del(.primaryDNS[])' -i /opt/overture/config.yml
  IFS=","
  for dns in $1; do
    name=$(echo ${dns} | cut -d "/" -f1) \
    address=$(echo ${dns} | cut -d "/" -f2) \
    protocol=$(echo ${dns} | cut -d "/" -f3) \
    yq eval '.primaryDNS += [{"name":strenv(name),"address":strenv(address),"protocol":strenv(protocol),"socks5Address":"","timeout":6,"ednsClientSubnet":{"policy":"auto","externalIP":"","noCookie":true}}]' \
      -i /opt/overture/config.yml
    yq eval -j /opt/overture/config.yml > /tmp/config.json
    yq eval -P /tmp/config.json > config.yml
    rm /tmp/config.json
  done
}

replaceAlternativeDns() {
  yq eval 'del(.alternativeDNS[])' -i /opt/overture/config.yml
  IFS=","
  for dns in $1; do
    name=$(echo ${dns} | cut -d "/" -f1) \
    address=$(echo ${dns} | cut -d "/" -f2) \
    protocol=$(echo ${dns} | cut -d "/" -f3) \
    yq eval '.alternativeDNS += [{"name":strenv(name),"address":strenv(address),"protocol":strenv(protocol),"socks5Address":"","timeout":6,"ednsClientSubnet":{"policy":"auto","externalIP":"","noCookie":true}}]' \
      -i /opt/overture/config.yml
    yq eval -j /opt/overture/config.yml > /tmp/config.json
    yq eval -P /tmp/config.json > config.yml
    rm /tmp/config.json
  done
}

if [ ! -z "${PRIMARY_DNS}" ]; then
  replacePrimaryDns "${PRIMARY_DNS}"
fi

if [ ! -z "${ALTERNATIVE_DNS}" ]; then
  replaceAlternativeDns "${ALTERNATIVE_DNS}"
fi

if [ ! -z "${IP_NETWORK_FILE_PRIMARY}" ]; then
  yq e '.ipNetworkFile.primary = "${IP_NETWORK_FILE_PRIMARY}"' -i /opt/overture/config.yml
fi

if [ ! -z "${DOMAIN_FILE_ALTERNATIVE}" ]; then
  yq e '.domainFile.alternative = "${DOMAIN_FILE_ALTERNATIVE}"' -i /opt/overture/config.yml
fi

if [ ! -z "${HOSTS_FILE}" ]; then
  yq e '.hostsFile.hostsFile = "${HOSTS_FILE}"' -i /opt/overture/config.yml
fi

if [ ! -z "${CACHE_SIZE}" ]; then
  yq e '.cacheSize = env(CACHE_SIZE)' -i /opt/overture/config.yml
fi

if [ ! -z "${ALTERNATIVE_DNS_SOCKS5_PROXY}" ]; then
  yq e '.alternativeDNS.[].socks5Address = "${ALTERNATIVE_DNS_SOCKS5_PROXY}"' -i /opt/overture/config.yml
fi

exec "$@"
