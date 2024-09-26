# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build
COPY . .

RUN go mod download

RUN go vet -v ./...
RUN go test -v ./...

RUN GOARCH=${TARGETOS} GOARCH=${TARGETARCH} go build -o ./tcplogger cmd/tcplogger/main.go

FROM gcr.io/distroless/base-nossl-debian12

COPY --from=builder /build/tcplogger /

CMD ["/tcplogger"]
