FROM golang:1.26-alpine AS builder

# Copy shared module so replace directive resolves
COPY gbs-common /gbs-common

WORKDIR /app
COPY gbs-cms-api/go.mod gbs-cms-api/go.sum ./
RUN go mod download

COPY gbs-cms-api/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gbs-cms-api ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /gbs-cms-api .

EXPOSE 8081
CMD ["./gbs-cms-api"]
