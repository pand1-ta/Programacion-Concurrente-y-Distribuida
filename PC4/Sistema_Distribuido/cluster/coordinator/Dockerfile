FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /coordinator ./cluster/coordinator/cmd/coordinator

FROM alpine:3.18
COPY --from=builder /coordinator /coordinator
EXPOSE 8081
ENTRYPOINT ["/coordinator"]