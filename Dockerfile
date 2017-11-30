FROM golang:1.9-alpine AS build-env
ADD . /src
RUN apk update && apk add git
RUN cd /src && go get github.com/miekg/dns && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fakedns .

FROM alpine
LABEL maintainer="pavel@evilmind.ru"
ENV DEFAULT_IPV4 127.0.0.1
ENV DEFAULT_IPV6 ::1
COPY --from=build-env /src/fakedns /usr/local/bin
EXPOSE 53
CMD fakedns

