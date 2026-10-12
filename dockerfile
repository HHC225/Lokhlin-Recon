# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:22-bookworm-slim AS frontend
WORKDIR /source/frontend
ENV HUSKY=0
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --ignore-scripts --no-audit --no-fund
COPY frontend/ ./
RUN npm run build:pro

FROM --platform=$BUILDPLATFORM golang:1.26.3-bookworm AS build
ARG TARGETARCH
WORKDIR /source
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
# Replace the old bundle before compiling the embedded frontend.
RUN rm -rf ./cmd/main/static && mkdir -p ./cmd/main/static
COPY --from=frontend /source/frontend/dist-pro/ ./cmd/main/static/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -p 2 -trimpath -ldflags "-s -w" -o /out/recon-server ./cmd/main

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git tzdata unzip \
    && rm -rf /var/lib/apt/lists/*
ENV TZ=UTC RECON_STATE_DIR=/opt/recon
WORKDIR /opt/recon
COPY --from=build /out/recon-server /usr/local/bin/recon-server
RUN mkdir -p files images uploads export
VOLUME ["/opt/recon"]
EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/recon-server"]
