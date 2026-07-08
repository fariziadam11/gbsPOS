FROM golang:1.26-alpine AS builder

COPY gbs-common /gbs-common

WORKDIR /app
COPY gbs-pos-api/go.mod gbs-pos-api/go.sum ./
RUN go mod download

COPY gbs-pos-api/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gbs-pos-api ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /gbs-pos-api ./gbs-pos-api

EXPOSE 8080
CMD ["./gbs-pos-api"]
