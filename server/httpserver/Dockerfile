# Tag: garukun/golgtm
FROM alpine

RUN apk add -q --update \
    && apk add -q \
            bash \
            curl \
    && rm -rf /var/cache/apk/*

COPY ./_out/app /usr/local/bin/app

EXPOSE 8080

ENTRYPOINT ["app"]
