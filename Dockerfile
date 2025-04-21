FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tldr-llm-telegram-bot ./cmd/tldr-llm-telegram-bot/main.go

FROM ubuntu:latest

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /root/

COPY --from=builder /app/tldr-llm-telegram-bot .

CMD ["./tldr-llm-telegram-bot"]