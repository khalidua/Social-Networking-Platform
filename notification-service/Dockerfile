FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server .
COPY migrations ./migrations
EXPOSE 8085
CMD ["./server"]
