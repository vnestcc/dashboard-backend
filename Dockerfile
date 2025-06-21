FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o server

FROM gcr.io/distroless/cc
COPY --from=builder /app/server /server
COPY --from=builder /app/config.toml /config.toml
EXPOSE 8080
CMD ["/server"]
