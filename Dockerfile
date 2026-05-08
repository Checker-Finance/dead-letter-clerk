# syntax=docker/dockerfile:1
FROM golang:1.25 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy && \
     go build -o dead-letter-clerk ./cmd/clerk

FROM gcr.io/distroless/base-debian11
WORKDIR /dead-letter-clerk
COPY --from=builder /app/dead-letter-clerk .
COPY config.yaml .

ENTRYPOINT ["./dead-letter-clerk"]