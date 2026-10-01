FROM golang:1.27.1-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod tidy && go mod verify
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/reforge-api .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/reforge-api /app/reforge-api
COPY --chown=65532:65532 content/.keep /var/lib/reforge/content/.keep
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/reforge-api"]
