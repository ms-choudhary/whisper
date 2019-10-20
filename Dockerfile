FROM golang:alpine as build
RUN apk update \
    && apk add git
COPY whisper.go .
RUN go get -d -v .
RUN go build -o /whisper whisper.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /whisper /usr/bin/whisper
CMD ["whisper"]
