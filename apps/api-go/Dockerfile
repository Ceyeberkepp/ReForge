FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/reforge-api .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/reforge-api /app/reforge-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/reforge-api"]
