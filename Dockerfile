FROM golang:alpine as builder

RUN apk add --no-cache git

RUN mkdir /build

ADD . /build/

WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o gowww .

FROM alpine:latest

RUN apk add tzdata curl

RUN mkdir -p /gowww

COPY --from=builder /build/gowww /gowww/

RUN mkdir -p /gowww/vhosts

VOLUME /gowww/vhosts

WORKDIR /gowww

ENV TZ Europe/London
ENV GOWWW_ROOT /gowww/vhosts
ENV GOWWW_PORT 8080

EXPOSE 8080/tcp

CMD ["./gowww"]