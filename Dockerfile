FROM golang:1.12

WORKDIR /go/src/app
COPY . .

RUN make plugin.auth
RUN make server

ENTRYPOINT ["./_output/bin/server"]
CMD ["s"]
