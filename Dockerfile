# syntax=docker/dockerfile:1

FROM golang:latest

WORKDIR /root

# Download Go modules
COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app
EXPOSE 9090

CMD ["/app", "-host", "0.0.0.0"]
