FROM golang:alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o lywsd03mmc-prom-expo main.go
FROM alpine
WORKDIR /app
COPY --from=builder /build/lywsd03mmc-prom-expo lywsd03mmc-prom-expo
ENTRYPOINT ["/app/lywsd03mmc-prom-expo"]