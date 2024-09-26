FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY . .

RUN go mod download

RUN go vet -v ./...
RUN go test -v ./...

RUN go build -o ./tcplogger cmd/tcplogger/main.go

FROM gcr.io/distroless/base-nossl-debian12

COPY --from=builder /build/tcplogger /

CMD ["/tcplogger"]
