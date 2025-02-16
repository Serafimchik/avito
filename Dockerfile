FROM golang:1.20.3-alpine AS builder

COPY . /github.com/Serafimchik/avito/source/
WORKDIR /github.com/Serafimchik/avito/source/

RUN go mod download
RUN go build -o ./bin/crud_server ./cmd/server/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/Serafimchik/avito/source/bin/crud_server .

CMD [ "./crud_server" ]