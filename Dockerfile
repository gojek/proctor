FROM golang:1.12 AS builder
WORKDIR /go/src/app
COPY . .
RUN make plugin.auth
RUN make server

FROM ubuntu:latest
RUN apt-get update
RUN apt-get install -y ca-certificates
WORKDIR /app/
COPY --from=builder /go/src/app/_output/bin/server .
COPY --from=builder /go/src/app/_output/bin/plugin/auth.so .
COPY --from=builder /go/src/app/migrations ./migrations
ENV PROCTOR_AUTH_PLUGIN_BINARY=/app/auth.so

ENTRYPOINT ["./server"]
CMD ["s"]
