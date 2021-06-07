FROM golang:1.15.2

WORKDIR go/src
COPY . .
ENV GO111MODULE=on \
    GORPOXY=https://goproxy.cn,direct

RUN go build -o server
RUN mv ./server /
