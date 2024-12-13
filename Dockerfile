FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

COPY vendor ./vendor

COPY . ./

RUN go build -mod=vendor -o bin/app cmd/main.go

FROM alpine:3.20    

WORKDIR /app

COPY --from=builder /app/bin/app ./app

EXPOSE 5000

CMD [ "./bin/app" ]