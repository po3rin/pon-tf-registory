# syntax=docker/dockerfile:1
FROM golang:1.19

RUN apt update && apt install gnupg
# RUN gpg --armor --export AAADatg43refg

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]