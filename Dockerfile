# syntax=docker/dockerfile:1

# Образ для прода: сервер, собранный интерфейс и записи из audio/. Админки и синтезатора
# здесь нет — озвучку выбирают локально (make api) и коммитят в audio/.
#
#   docker build -t domashka .                              # под текущую платформу
#   docker buildx build --platform linux/amd64 -t domashka .  # под обычный сервер с Mac

# Сборочные стадии идут на родной платформе машины: результат у них переносимый.
FROM --platform=$BUILDPLATFORM node:26-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS server
WORKDIR /src/server
COPY server/ ./
# Тесты проверяют page.json и speech.json: с битым контентом образ не соберётся.
RUN CGO_ENABLED=0 go test ./...
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/domashka .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=server /out/domashka ./domashka
COPY --from=web /src/web/dist ./web
COPY audio ./audio
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD ["/app/domashka", "-healthcheck", "-addr", "0.0.0.0:8080"]
ENTRYPOINT ["/app/domashka"]
CMD ["-addr", "0.0.0.0:8080", "-audio", "/app/audio", "-static", "/app/web"]
