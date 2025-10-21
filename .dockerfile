FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /rate-limiter ./cmd/server

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /rate-limiter .

EXPOSE 8080

CMD ["./rate-limiter"]