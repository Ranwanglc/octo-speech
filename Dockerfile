FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o octo-speech ./cmd/speech

FROM debian:bookworm-slim
RUN useradd -u 10001 -r -s /bin/false appuser
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/octo-speech /usr/local/bin/
USER 10001
ENTRYPOINT ["octo-speech"]
