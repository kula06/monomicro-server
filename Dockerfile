FROM golang:1.21-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/monomicro-server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
	&& addgroup -S app \
	&& adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/monomicro-server /app/monomicro-server

USER app
EXPOSE 8080

ENTRYPOINT ["/app/monomicro-server"]
