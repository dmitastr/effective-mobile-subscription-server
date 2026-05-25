FROM golang:1.25 AS builder
WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o subscription-server ./cmd


FROM alpine
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/subscription-server /app/subscription-server
COPY migrations/ /app/migrations/
RUN touch .env


EXPOSE 8080
CMD ["/app/subscription-server"]