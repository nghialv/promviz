FROM alpine:3.20

RUN apk --update add ca-certificates atop

ADD build/promviz /bin

EXPOSE 9091

ENTRYPOINT ["/bin/promviz"]
