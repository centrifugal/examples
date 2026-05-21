FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o backend .

FROM alpine:latest
COPY --from=builder /app/backend /backend
EXPOSE 3001
CMD ["/backend"]
