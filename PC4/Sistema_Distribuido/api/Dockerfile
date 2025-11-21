FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/api-server ./cmd/api

FROM alpine:3.18
RUN apk add --no-cache ca-certificates

# AHORA ESTABLECEMOS EL WORKDIR REAL
WORKDIR /app

# Copiamos el binario
COPY --from=builder /app/api-server /app/api-server

# Copiamos el dataset (será sobreescrito por el volumen si existe)
COPY ./dataset /app/dataset

EXPOSE 8080
ENTRYPOINT ["./api-server"]