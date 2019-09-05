FROM golang:1.12 AS builder
WORKDIR /go/src/app
COPY . .
RUN make plugin.auth
RUN make server

FROM ubuntu:latest
WORKDIR /app/
COPY --from=builder /go/src/app/_output/bin/server .
COPY --from=builder /go/src/app/_output/bin/plugin/auth.so .
ENV PROCTOR_AUTH_PLUGIN_BINARY=/app/auth.so

ENTRYPOINT ["./server"]
CMD ["s"]
