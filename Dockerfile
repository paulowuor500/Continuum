FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
COPY apps/api ./
RUN go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/server /app/server
EXPOSE 8080
CMD ["/app/server"]
