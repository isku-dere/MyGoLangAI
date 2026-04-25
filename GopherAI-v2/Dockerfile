FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY . .
RUN go build -o /out/gopherai ./main.go

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/gopherai /app/gopherai
COPY config/config.docker.toml /app/config/config.toml

EXPOSE 9090

CMD ["/app/gopherai"]

