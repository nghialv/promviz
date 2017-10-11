FROM golang:1.8-alpine3.5

ADD main.go /
ADD metrics.txt /

EXPOSE 30001

ENTRYPOINT ["go","run","/main.go"]
