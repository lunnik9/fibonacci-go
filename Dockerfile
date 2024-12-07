FROM golang:1.23.3-alpine as builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fibonacci ./cmd/main.go

FROM alpine:3.18

WORKDIR /root/
COPY --from=builder /app/fibonacci .

EXPOSE 50051 9090
ENTRYPOINT ["./fibonacci"]
