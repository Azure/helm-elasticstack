FROM alpine:3.6

ARG VERSION=5.44

ENV STUNNEL_URL https://www.stunnel.org/downloads/stunnel-$VERSION.tar.gz

ENV STUNNEL_FILE stunnel-$VERSION.tar.gz
ENV STUNNEL_TEMP stunnel-$VERSION-build
ENV STUNNEL_DEPS openssl
ENV BUILD_DEPS curl alpine-sdk openssl-dev

RUN set -xe \
    && apk update \
    && apk add $STUNNEL_DEPS $BUILD_DEPS \
    && mkdir $STUNNEL_TEMP \
        && cd $STUNNEL_TEMP \
        && curl -sSL $STUNNEL_URL -o $STUNNEL_FILE \
        && tar -xf $STUNNEL_FILE --strip 1 \
        && ./configure \
        && make install \
        && cd .. \
        && rm -rf $STUNNEL_TEMP $STUNNEL_FILE \
    && apk --purge del $BUILD_DEPS \
    && mkdir -p /etc/stunnel/

WORKDIR /etc/stunnel

CMD ["stunnel", "/config/stunnel.conf"]
