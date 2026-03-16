FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY cmd/ cmd/
COPY internal/ internal/

RUN go build -o server ./cmd/server

FROM alpine:3.21

WORKDIR /app
COPY --from=builder /app/server .
RUN mkdir -p event_logs

EXPOSE 8080
CMD ["./server"]
