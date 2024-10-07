# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build
COPY . .

RUN go mod download && \
  go vet -v ./... && \
  go test -v ./... && \
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 \
  go build -o ./tcplogger cmd/tcplogger/main.go

FROM gcr.io/distroless/base-nossl-debian12:nonroot

COPY --from=builder /build/tcplogger /

CMD ["/tcplogger"]
