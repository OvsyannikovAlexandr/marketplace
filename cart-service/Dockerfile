FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o cart-service ./cmd/main.go

EXPOSE 8084

CMD [ "./cart-service" ]