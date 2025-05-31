FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=0
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o server

FROM alpine:latest AS runtime
COPY --from=builder /app/server /root/server
COPY ./config.toml /root/config.toml
CMD ["/root/server"]
