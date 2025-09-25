FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=dev
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.Version=${VERSION} -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.Commit=${COMMIT} -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.BuildDate=${BUILD_DATE}" -o /out/flight-booking ./cmd/flight-booking

FROM alpine:3.19
RUN adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=builder /out/flight-booking /app/flight-booking
ENTRYPOINT ["/app/flight-booking"]
