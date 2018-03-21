FROM alpine:3.6

ENV OAUTH2_PROXY_VERSION 2.2
ENV OAUTH2_PROXY_PACAKGE ${OAUTH2_PROXY_VERSION}.0.linux-amd64
ENV GO_VERSION 1.8.1

RUN apk --update add curl

RUN curl -sL -o oauth2_proxy.tar.gz \
    "https://github.com/bitly/oauth2_proxy/releases/download/v$OAUTH2_PROXY_VERSION/oauth2_proxy-$OAUTH2_PROXY_PACAKGE.go$GO_VERSION.tar.gz" \
  && tar xzvf oauth2_proxy.tar.gz \
  && mv oauth2_proxy-$OAUTH2_PROXY_PACAKGE.go$GO_VERSION/oauth2_proxy /bin/ \
  && chmod +x /bin/oauth2_proxy \
  && rm -r oauth2_proxy*

ENTRYPOINT ["oauth2_proxy"]
