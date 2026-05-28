ARG GO_VERSION=1.26

FROM golang:${GO_VERSION} AS builder

WORKDIR /src

COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download

COPY . .

RUN wget "https://arms-apm-cn-hangzhou.oss-cn-hangzhou.aliyuncs.com/instgo/instgo-linux-amd64" -O instgo

RUN chmod +x instgo

RUN ./instgo go build -o /out/jcourse-api ./cmd/api

FROM alpine:3.22

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/jcourse-api /usr/local/bin/jcourse-api

COPY config/config.example.yaml /app/config/config.yaml

USER app

EXPOSE 8080

ENTRYPOINT ["jcourse-api"]
