FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gbs-pos-api ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /gbs-pos-api .

EXPOSE 8080
CMD ["./gbs-pos-api"]
