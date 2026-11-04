
# Stage 1: Build
FROM golang:1.27.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o autocomplete .

# Stage 2: Runtime
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/autocomplete ./autocomplete

EXPOSE 8080

CMD ["./autocomplete"]
