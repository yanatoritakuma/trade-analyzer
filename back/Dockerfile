# back/Dockerfile（本番に近い通常ビルド・マルチステージ）
FROM golang:1.25 AS builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o main .

FROM alpine AS dev
WORKDIR /app
COPY --from=builder /app/main /app/main
EXPOSE 8080

CMD ["./main"]
