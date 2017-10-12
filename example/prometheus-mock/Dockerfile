FROM golang:1.8-alpine3.5

EXPOSE 30001

RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

RUN mkdir -p $GOPATH/src/app
ADD . $GOPATH/src/app

WORKDIR $GOPATH/src/app
RUN go build -o demo .

ENTRYPOINT ["/go/src/app/demo"]
