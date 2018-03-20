FROM alpine:3.5

RUN apk add --no-cache bash

WORKDIR /say-hello-world

COPY say_hello_world.sh /say-hello-world

ENTRYPOINT ./say_hello_world.sh
