# scorecard-server — hosted scorecard with SQLite-backed history, for k3s.
#
# Build:  ./build.sh
# Test:   podman run --rm -p 8080:8080 -v /tmp/scores:/data ghcr.io/jeks313/scorecard:latest
FROM docker.io/library/golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
# modernc.org/sqlite is pure Go, so CGO_ENABLED=0 works fine and keeps the
# final image tiny (no libc needed for the sqlite driver itself).
RUN CGO_ENABLED=0 go build -o /out/scorecard-server .

FROM gcr.io/distroless/static-debian12
COPY --from=build --chmod=755 /out/scorecard-server /app/scorecard-server
# --chmod: the source files are 640 on disk (no "other" read bit), and the
# pod runs as uid 1000 while COPY lands them owned by root -- without this,
# the server can't open them, and Go's FileServer silently falls back to a
# directory listing instead of erroring.
COPY --chmod=644 web/ /app/web/
WORKDIR /app
ENV WEB_DIR=/app/web \
    DB_PATH=/data/scores.db \
    PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/scorecard-server"]
