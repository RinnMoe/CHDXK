ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download

COPY . .

RUN go build -trimpath -ldflags="-s -w" -o /out/jcourse-taskworker ./cmd/taskworker

FROM alpine:3.22

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/jcourse-taskworker /usr/local/bin/jcourse-taskworker

COPY config/config.example.yaml /app/config/config.yaml

USER app

ENTRYPOINT ["jcourse-taskworker"]
