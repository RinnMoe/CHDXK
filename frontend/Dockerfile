# syntax=docker/dockerfile:1

ARG NODE_VERSION=24

FROM node:${NODE_VERSION}-alpine AS builder

WORKDIR /app

ARG VITE_SITE_NAME=CHDXK
ARG VITE_SUPPORT_EMAIL=""
ENV VITE_SITE_NAME=${VITE_SITE_NAME}
ENV VITE_SUPPORT_EMAIL=${VITE_SUPPORT_EMAIL}

RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

RUN pnpm install --frozen-lockfile

COPY . .

RUN pnpm build

FROM nginx:1.31-alpine AS runtime

COPY nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
