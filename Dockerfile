# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM node:26-alpine@sha256:40cd011b55415bc894e45b02fdd5a5dfde056921b7d23a64c74f80497af48797 AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --ignore-scripts
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY content/ ./content/
COPY config/ ./config/
RUN CGO_ENABLED=0 go test ./...
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/domashka ./cmd/domashka

FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99
WORKDIR /app
COPY --from=server /out/domashka ./domashka
COPY --from=web /src/web/dist ./web
COPY audio ./audio
COPY content ./content
COPY config/base.yaml config/production.yaml ./config/
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD ["/app/domashka", "-config", "/app/config/base.yaml", "-config", "/app/config/production.yaml", "-healthcheck"]
ENTRYPOINT ["/app/domashka", "-config", "/app/config/base.yaml", "-config", "/app/config/production.yaml"]
