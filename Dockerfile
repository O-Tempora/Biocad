FROM golang:alpine AS builder
WORKDIR /app
COPY . ./
RUN make build
FROM alpine
WORKDIR /app
COPY --from=builder /app/app ./
COPY --from=builder /app/config /app/config
COPY --from=builder /app/fonts /app/fonts
COPY --from=builder /app/Makefile ./