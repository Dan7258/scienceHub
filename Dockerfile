FROM golang:1.21.0 AS builder

WORKDIR /app

RUN go install github.com/revel/cmd/revel@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .


FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libreoffice-core \
        libreoffice-writer \
        libreoffice-calc \
        libreoffice-common \
        fonts-dejavu-core \
        ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

COPY --from=builder /go/bin/revel /usr/local/bin/revel

COPY --from=builder /app /app

EXPOSE 9000

CMD ["revel", "run", "."]