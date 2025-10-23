FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o rate-limiter ./cmd/server

EXPOSE 8080

CMD ["./rate-limiter"]