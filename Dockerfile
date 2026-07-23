FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o urlshortener ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
RUN mkdir -p /app/data

COPY --from=builder /app/urlshortener .

EXPOSE 8080
CMD ["./urlshortener"]