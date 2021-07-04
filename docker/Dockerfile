FROM alpine:latest
MAINTAINER SgrAlpha <admin@mail.sgr.io>

ARG OVERTURE_VERSION=1.7
ARG OVERTURE_ASSET=overture-linux-amd64

EXPOSE 53
WORKDIR /opt/overture
ENTRYPOINT ["/docker-entrypoint.sh"]

RUN apk --update --no-cache add curl yq && \
    curl -Ls https://github.com/shawn1m/overture/releases/download/v"${OVERTURE_VERSION}"/"${OVERTURE_ASSET}.zip" -o /tmp/"${OVERTURE_ASSET}.zip" && \
    unzip /tmp/"${OVERTURE_ASSET}.zip" -d /opt/overture && \
    ln -sf /opt/overture/"${OVERTURE_ASSET}" /opt/overture/overture && \
    curl https://raw.githubusercontent.com/17mon/china_ip_list/master/china_ip_list.txt \
        | tee /opt/overture/china_ip_list.txt && \
    curl https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt \
        | base64 -d \
        | sort -u \
        | sed '/^$\|@@/d' \
        | sed 's#!.\+##; s#|##g; s#@##g; s#http:\/\/##; s#https:\/\/##;' \
        | sed '/\*/d /apple\.com/d; /sina\.cn/d; /sina\.com\.cn/d; /baidu\.com/d; /qq\.com/d' \
        | sed '/^[0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+$/d' \
        | grep '^[0-9a-zA-Z\.-]\+$' | grep '\.' \
        | sed 's#^\.\+##' | sort -u \
        | tee /opt/overture/gfw_domain_list.txt && \
    yq e '.ipNetworkFile.primary = "./china_ip_list.txt"' -i /opt/overture/config.yml && \
    yq e '.domainFile.alternative = "./gfw_domain_list.txt"' -i /opt/overture/config.yml && \
    rm -rf /tmp/* /var/cache/apk/*

COPY ./docker-entrypoint.sh /

CMD /opt/overture/overture -c /opt/overture/config.yml
