FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o user-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/user-service .
COPY .env .

EXPOSE 8080

CMD ["./user-service"]