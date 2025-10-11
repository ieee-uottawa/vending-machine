FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /vend-server ./cmd/server/main.go

FROM scratch

COPY --from=builder /vend-server /vend-server

EXPOSE 3000

CMD ["./vend-server"]