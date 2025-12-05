# cluster/workers/Dockerfile

FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copia todo el código del proyecto al contenedor
COPY ../../.. .

# Compila el binario del worker
RUN go build -o worker ./cluster/workers/main.go

# Imagen final minimalista
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/worker .
EXPOSE 9000
CMD ["./worker"]
