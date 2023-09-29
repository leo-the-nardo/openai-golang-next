FROM golang:1.21.5-bullseye

WORKDIR /go/src

RUN ln -sf /bin/bash /bin/sh
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

CMD ["tail", "-f", "/dev/null"]