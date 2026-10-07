ARG GO_VERSION=1.26

FROM golang:${GO_VERSION} AS builder

WORKDIR /src

COPY backend/go.mod backend/go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 go build -trimpath -o /out/jcourse-api ./cmd/api

FROM node:22-bookworm-slim AS collector

ARG PNPM_VERSION=10.33.3
WORKDIR /collector
RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate
COPY tools/bkjw-course-source/package.json tools/bkjw-course-source/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY tools/bkjw-course-source/ ./
RUN pnpm build && pnpm prune --prod

FROM node:22-bookworm-slim

ENV PLAYWRIGHT_BROWSERS_PATH=/opt/playwright \
    JCOURSE_COURSE_SOURCE_ENABLED=true \
    JCOURSE_COURSE_SOURCE_NODE=/usr/local/bin/node \
    JCOURSE_COURSE_SOURCE_RUNNER=/app/tools/bkjw-course-source/dist/admin-runner.js \
    JCOURSE_COURSE_SOURCE_BROWSER_PATH=/opt/playwright

COPY --from=collector /collector /app/tools/bkjw-course-source
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && /app/tools/bkjw-course-source/node_modules/.bin/playwright install --with-deps chromium \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /app/.local/course-sync \
    && chown -R node:node /app/.local

WORKDIR /app

COPY --from=builder /out/jcourse-api /usr/local/bin/jcourse-api

COPY backend/config/config.example.yaml /app/config/config.yaml

USER node

EXPOSE 8080

ENTRYPOINT ["jcourse-api"]
