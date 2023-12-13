FROM golang:1.20-alpine as builder

WORKDIR /app

COPY . /app
RUN go mod tidy
RUN go build -o /app main.go

FROM alpine:3
WORKDIR /app

COPY config.yaml .
COPY private.pem .


COPY --from=builder /app/db /db
COPY --from=builder /app/main /app
