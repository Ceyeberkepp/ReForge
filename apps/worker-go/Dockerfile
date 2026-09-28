FROM golang:1.27.1-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/reforge-worker .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/reforge-worker /app/reforge-worker
USER nonroot:nonroot
ENTRYPOINT ["/app/reforge-worker"]
