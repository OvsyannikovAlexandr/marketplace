FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o product-service ./cmd/main.go

EXPOSE 8082

CMD [ "./product-service" ]