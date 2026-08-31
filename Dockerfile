# Stage 1: build the frontend
FROM node:22-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

# Stage 2: build the Go binary
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/ui/dist ./ui/dist
RUN go build -o bin/server ./cmd/server

# Stage 3: minimal runtime image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=go-builder /app/bin/server ./server

ENV APP_ENV=production
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["./server"]
