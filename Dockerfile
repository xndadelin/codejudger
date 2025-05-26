FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

FROM debian:bookworm

WORKDIR /root/

RUN apt-get update && apt-get install -y curl apt-utils gnupg

RUN mkdir -p /etc/apt/keyrings && \
    curl https://www.ucw.cz/isolate/debian/signing-key.asc >/etc/apt/keyrings/isolate.asc && \
    echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/isolate.asc] http://www.ucw.cz/isolate/debian/ bookworm-isolate main" >> /etc/apt/sources.list

RUN apt-get update && apt-get install -y isolate

COPY --from=builder /app/server .

CMD ["./server"]