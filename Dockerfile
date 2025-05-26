FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

FROM debian:bookworm

WORKDIR /root/

RUN apt-get update && apt-get install -y \
    curl \
    apt-utils \
    gnupg \
    build-essential \
    pkg-config \
    libcap-dev \
    libsystemd-dev \
    git \
    make

RUN git clone https://github.com/ioi/isolate.git /tmp/isolate && \
    cd /tmp/isolate && \
    make isolate && \
    cp isolate /usr/local/bin/ && \
    cd / && rm -rf /tmp/isolate

RUN mkdir -p /var/lib/isolate && chmod 777 /var/lib/isolate

COPY --from=builder /app/server .

CMD ["./server"]