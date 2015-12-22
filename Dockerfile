FROM scratch

ADD _out/golgtm /golgtm

ENTRYPOINT ["/golgtm"]
