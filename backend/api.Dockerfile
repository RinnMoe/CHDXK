ARG GO_VERSION=1.26
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -trimpath -o /out/jcourse-api ./cmd/api

FROM alpine:3.22
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app \
    && mkdir -p /app/.local/course-cache && chown -R app:app /app/.local
WORKDIR /app
ENV JCOURSE_COURSE_SOURCE_ENABLED=true
COPY --from=builder /out/jcourse-api /usr/local/bin/jcourse-api
COPY backend/config/config.example.yaml /app/config/config.yaml
USER app
EXPOSE 8080
ENTRYPOINT ["jcourse-api"]
