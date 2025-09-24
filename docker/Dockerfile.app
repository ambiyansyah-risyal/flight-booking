FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /out/flight-booking ./cmd/flight-booking

FROM alpine:3.19
RUN adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=builder /out/flight-booking /app/flight-booking
ENTRYPOINT ["/app/flight-booking"]

