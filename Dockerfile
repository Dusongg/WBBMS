FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bookadmin .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /app/bookadmin /app/bookadmin
COPY config.yaml /app/config.yaml

EXPOSE 8888

ENTRYPOINT ["/app/bookadmin"]
CMD ["-mode=api"]
