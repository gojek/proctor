FROM golang:1.12 AS builder
WORKDIR /go/src/app
COPY . .
RUN make plugin.auth
RUN make plugin.slack
RUN make server

FROM ubuntu:latest
RUN apt-get update
RUN apt-get install -y ca-certificates
WORKDIR /app/
COPY --from=builder /go/src/app/_output/bin/server .
COPY --from=builder /go/src/app/_output/bin/plugin/auth.so .
COPY --from=builder /go/src/app/_output/bin/plugin/slack.so .
COPY --from=builder /go/src/app/migrations ./migrations

ENTRYPOINT ["./server"]
CMD ["s"]
