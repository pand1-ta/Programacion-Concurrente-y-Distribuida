FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copiamos los archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del proyecto
COPY . .

# Compilamos la API
RUN CGO_ENABLED=0 go build -o /app/api-server ./api/cmd/api

# Copiamos la documentación Swagger
COPY api/docs /app/api/docs

# Copiamos el dataset también en el builder
RUN mkdir -p /app/dataset
COPY api/dataset/movies.csv /app/dataset/movies.csv
COPY api/dataset/usuarios_mapping.csv /app/dataset/usuarios_mapping.csv
COPY api/dataset/ratings.csv /app/dataset/ratings.csv
COPY api/dataset/peliculas_mapping.csv /app/dataset/peliculas_mapping.csv
COPY api/dataset/matriz_usuarios_peliculas.csv /app/dataset/matriz_usuarios_peliculas.csv

# -------------------
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copiamos el binario desde el builder
COPY --from=builder /app/api-server /app/api-server

# Copiamos también el dataset y la documentación Swagger
COPY --from=builder /app/dataset /app/dataset
COPY --from=builder /app/api/docs /app/api/docs

EXPOSE 8080
CMD ["./api-server"]
