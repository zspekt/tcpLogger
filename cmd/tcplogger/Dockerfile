FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY . .

RUN go mod download

RUN go vet -v ./...
RUN go test -v ./...

RUN go build -o ./tcplogger

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/tcplogger /

CMD ["/tcplogger"]


