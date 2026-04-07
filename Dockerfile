FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

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
COPY --from=frontend /app/web/dist /app/web/dist
WORKDIR /app
ENTRYPOINT ["/bin/server"]

FROM alpine:3.19 AS worker
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/worker /bin/worker
ENTRYPOINT ["/bin/worker"]
