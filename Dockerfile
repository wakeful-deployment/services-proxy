FROM buildpack-deps:jessie-curl

MAINTAINER Nathan Herald <nathan.herald@microsoft.com>

RUN apt-get update \
 && mkdir /opt/app \
 && mkdir /opt/app/bin \
 && mkdir /opt/src

ENV PATH /opt/app/bin:$PATH

# consul-template

ENV CONSUL_TEMPLATE_VERSION 0.12.2
ENV CONSUL_TEMPLATE_DOWNLOAD_URL https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip
ENV CONSUL_TEMPLATE_SHA256 a8780f365bf5bfad47272e4682636084a7475ce74b336cdca87c48a06dd8a193

RUN apt-get install unzip -y

RUN curl -o /opt/src/consul-template.zip "${CONSUL_TEMPLATE_DOWNLOAD_URL}" \
 && echo "${CONSUL_TEMPLATE_SHA256}  /opt/src/consul-template.zip" > /opt/src/consul-template.sha256 \
 && sha256sum -c /opt/src/consul-template.sha256 \
 && cd /opt/app/bin \
 && unzip /opt/src/consul-template.zip \
 && chmod +x /opt/app/bin/consul-template \
 && rm /opt/src/consul-template.zip

# haproxy

RUN apt-get install -y --no-install-recommends libssl1.0.0 libpcre3

ENV HAPROXY_MAJOR 1.6
ENV HAPROXY_VERSION 1.6.3
ENV HAPROXY_DOWNLOAD_URL http://www.haproxy.org/download/${HAPROXY_MAJOR}/src/haproxy-${HAPROXY_VERSION}.tar.gz
ENV HAPROXY_MD5 3362d1e268c78155c2474cb73e7f03f9

RUN curl -SL -o haproxy.tar.gz "${HAPROXY_DOWNLOAD_URL}" \
 && echo "${HAPROXY_MD5}  haproxy.tar.gz" | md5sum -c \
 && mkdir -p /opt/src/haproxy \
 && tar -xzf haproxy.tar.gz -C /usr/src/haproxy --strip-components=1 \
 && rm haproxy.tar.gz \
 && make -C /opt/src/haproxy \
      TARGET=linux2628 \
      USE_PCRE=1 \
      PCREDIR= \
      USE_OPENSSL=1 \
      USE_ZLIB=1 \
      all \
      install-bin \
 && rm -rf /opt/src/haproxy

# go

RUN apt-get install -y --no-install-recommends g++ gcc libc6-dev make

ENV GOLANG_VERSION 1.5.2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA1 cae87ed095e8d94a81871281d35da7829bd1234e

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
 && echo "$GOLANG_DOWNLOAD_SHA1  golang.tar.gz" | sha1sum -c - \
 && tar -C /usr/local -xzf golang.tar.gz \
 && rm golang.tar.gz

ENV GOPATH /opt
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# move configs and stuff
ADD . /opt/app

# move go code into a place that can be compiled
ADD . /opt/src/github.com/wakeful-deployment/services-proxy

# compile
RUN cd /opt/src/github.com/wakeful-deployment/services-proxy \
 && go build \
 && mv ./services-proxy /opt/app/bin/

# always...

RUN chmod +x /opt/app/bin/*

ARG sha
ARG start

RUN echo $start > /opt/start \
 && chmod +x /opt/start

RUN echo $sha > /opt/app/sha

CMD /opt/start

