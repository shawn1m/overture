#!/bin/sh
set -e

replacePrimaryDns() {
  yq eval 'del(.primaryDNS[])' -i /opt/overture/config.yml
  IFS=","
  for dns in $1; do
    name=$(echo ${dns} | cut -d "/" -f1) \
    protocol=$(echo ${dns} | rev | cut -d "/" -f1 | rev) \
    address=$(echo ${dns} | cut -d "/" -f2- | rev | cut -d "/" -f2- | rev) \
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
    protocol=$(echo ${dns} | rev | cut -d "/" -f1 | rev) \
    address=$(echo ${dns} | cut -d "/" -f2- | rev | cut -d "/" -f2- | rev) \
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
  yq eval '.ipNetworkFile.primary = strenv(IP_NETWORK_FILE_PRIMARY)' -i /opt/overture/config.yml
fi

if [ ! -z "${DOMAIN_FILE_ALTERNATIVE}" ]; then
  yq eval '.domainFile.alternative = strenv(DOMAIN_FILE_ALTERNATIVE)' -i /opt/overture/config.yml
fi

if [ ! -z "${HOSTS_FILE}" ]; then
  yq eval '.hostsFile.hostsFile = strenv(HOSTS_FILE)"' -i /opt/overture/config.yml
fi

if [ ! -z "${CACHE_SIZE}" ]; then
  yq eval '.cacheSize = env(CACHE_SIZE)' -i /opt/overture/config.yml
fi

if [ ! -z "${ALTERNATIVE_DNS_SOCKS5_PROXY}" ]; then
  yq eval '.alternativeDNS.[].socks5Address = strenv(ALTERNATIVE_DNS_SOCKS5_PROXY)' -i /opt/overture/config.yml
fi

exec "$@"
