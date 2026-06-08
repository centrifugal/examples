FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /server .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=build /server /usr/local/bin/server
EXPOSE 5000
CMD ["server"]
