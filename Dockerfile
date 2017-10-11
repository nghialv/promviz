FROM alpine:3.5

RUN apk --update add ca-certificates atop

ADD build/promviz /

ENTRYPOINT ["/promviz"]