FROM golang:1.21.3 as builder

WORKDIR /app

COPY . .

WORKDIR /app/web-server

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest

COPY --from=builder /app/web-server/app /app

COPY --from=builder /app/web-server/config.json /config.json

ENTRYPOINT ["/app"]
