FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o backend .

FROM alpine:latest
COPY --from=builder /app/backend /backend
CMD ["/backend"]
