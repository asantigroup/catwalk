FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o catwalk .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/catwalk /usr/bin/catwalk

EXPOSE 8080

USER nobody

CMD ["/usr/bin/catwalk"]
