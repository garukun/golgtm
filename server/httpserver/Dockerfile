# Tag: garukun/golgtm
FROM alpine

RUN apk add -q --update \
    && apk add -q \
            bash \
            curl \
    && rm -rf /var/cache/apk/*

COPY ./_out/golgtm /usr/local/bin/golgtm

EXPOSE 8080

ENTRYPOINT ["golgtm"]
