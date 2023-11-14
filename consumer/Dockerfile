FROM golang:1.21.3 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer .

FROM alpine:latest  

COPY --from=builder /app/consumer /consumer

COPY ./config.json ./config.json

ENTRYPOINT ["/consumer"]
