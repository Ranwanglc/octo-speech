FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o octo-speech ./cmd/speech

FROM debian:bookworm-slim
COPY --from=builder /app/octo-speech /usr/local/bin/
ENTRYPOINT ["octo-speech"]
