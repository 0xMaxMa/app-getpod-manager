FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN VERSION=$(grep '^version:' app.yaml | awk '{print $2}') && \
    go build -ldflags "-X main.Version=${VERSION:-dev}" -o getpod-manager .

FROM alpine:3.19
RUN apk add --no-cache wget
WORKDIR /app
COPY --from=builder /app/getpod-manager .
COPY scripts/ scripts/
EXPOSE 5990
CMD ["./getpod-manager"]
