FROM golang:1.21.3 as builder

WORKDIR /app

COPY . .

WORKDIR /app/consumer

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest

COPY --from=builder /app/consumer/app /app

COPY --from=builder /app/consumer/config.json /config.json

ENTRYPOINT ["/app"]