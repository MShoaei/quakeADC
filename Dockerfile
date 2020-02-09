FROM golang:1.13-alpine as builder
WORKDIR /app
RUN apk add --no-cache build-base libpcap-dev
COPY go.* ./
RUN go mod download
COPY . .
#RUN go build -ldflags '-extldflags "-static"' -o qADC ./rpiCMD/main.go
RUN go build -o qADC ./rpiCMD/main.go

FROM alpine:latest
RUN apk update && apk add libpcap-dev
COPY --from=builder /app/qADC /usr/bin/qADC
ENTRYPOINT ["/usr/bin/qADC"]
EXPOSE 9090
CMD ["server"]