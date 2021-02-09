FROM golang:1.15-alpine as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
RUN apk update 
RUN apk add --no-cache gcc libusb-dev musl-dev
COPY . .
#RUN go build -ldflags '-extldflags "-static"' -o qADC ./rpiCMD/main.go
RUN go build -o quakeBinary ./rpiCMD/

FROM alpine:latest
RUN apk update && apk add libusb-dev
COPY --from=builder /app/quakeBinary /usr/bin/rpiCMD
RUN chmod +x /usr/bin/rpiCMD
ENV PORT=9090
ENTRYPOINT ["/usr/bin/rpiCMD"]
EXPOSE 9090
CMD ["server", "--skip"]
