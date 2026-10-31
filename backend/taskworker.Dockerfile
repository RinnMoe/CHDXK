ARG GO_VERSION=1.26

FROM golang:${GO_VERSION} AS builder

WORKDIR /src

COPY backend/go.mod backend/go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 go build -trimpath -o /out/jcourse-taskworker ./cmd/taskworker

FROM alpine:3.22

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/.local/course-cache /app/.local/canteen-media \
    && chown -R app:app /app/.local

WORKDIR /app

COPY --from=builder /out/jcourse-taskworker /usr/local/bin/jcourse-taskworker

COPY backend/config/config.example.yaml /app/config/config.yaml

USER app

ENTRYPOINT ["jcourse-taskworker"]
