FROM golang:alpine as build

WORKDIR /temp/build
COPY . .

RUN apk --no-cache add build-base git && make clean build

FROM alpine:latest

WORKDIR /opt/tinycloud
COPY --from=build /temp/build/target/tinycloud /opt/tinycloud/bin/tinycloud

RUN apk --no-cache add ca-certificates

VOLUME [ "/opt/tinycloud/var" ]

ENV TINYCLOUD_DATA /opt/tinycloud/var
ENV TINYCLOUD_ADDRESS :80
ENV PATH /opt/tinycloud/bin:${PATH}

EXPOSE 80/tcp

CMD [ "tinycloud", "start" ]
