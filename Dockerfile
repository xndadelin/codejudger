FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

FROM debian:bookworm

WORKDIR /root/

RUN apt-get update && apt-get install -y \
    curl \
    git \
    g++ \
    gcc \
    make \
    build-essential \
    pkg-config \
    libcap-dev \
    libsystemd-dev \
    ca-certificates \
    python3

RUN git clone https://github.com/ioi/isolate.git && \
    cd isolate && \
    make isolate && \
    make install && \
    cd .. && rm -rf isolate

RUN mkdir -p /var/lib/isolate && chmod 777 /var/lib/isolate

COPY --from=builder /app/server .

CMD ["./server"]
