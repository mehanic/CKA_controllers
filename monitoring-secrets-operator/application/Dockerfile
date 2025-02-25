# # Build stage
# FROM golang:1.21 AS builder
# WORKDIR /app
# COPY main.go .
# RUN go mod init webapp && go mod tidy && go build -o webapp

# # Final image
# FROM alpine:latest
# RUN apk add --no-cache ca-certificates
# WORKDIR /root/
# COPY --from=builder /app/webapp .
# CMD ["./webapp"]

# commited application
# FROM golang:1.21
# WORKDIR /app
# COPY go.mod ./
# RUN go mod download
# COPY *.go ./
# RUN go build -o /go-podman-demo
# EXPOSE 8080
# CMD [ "/go-podman-demo" ]

FROM golang:1.21
WORKDIR /app
COPY go.mod ./
RUN go mod download
# Add go get to ensure go.sum is populated correctly
RUN go get github.com/prometheus/client_golang/prometheus github.com/prometheus/client_golang/prometheus/promhttp
COPY *.go ./
RUN go build -o /go-podman-demo
EXPOSE 8080
CMD [ "/go-podman-demo" ]
