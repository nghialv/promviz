FROM alpine:3.5

RUN apk --update add ca-certificates atop

ADD build/k8s-config-reloader /

ENTRYPOINT ["/k8s-config-reloader"]