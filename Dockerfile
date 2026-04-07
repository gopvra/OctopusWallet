FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /bin/worker ./cmd/worker

FROM alpine:3.19 AS server
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/server /bin/server
ENTRYPOINT ["/bin/server"]

FROM alpine:3.19 AS worker
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/worker /bin/worker
ENTRYPOINT ["/bin/worker"]
