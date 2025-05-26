
FROM golang:1.22-alpine AS builder
# alpine - легковесная версия Linux --> образы меньше

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -v -o myapp ./main.go


FROM alpine:latest

WORKDIR /root/


COPY --from=builder /app/myapp .
COPY --from=builder /app/static ./static

EXPOSE 8080


CMD ["./myapp"]