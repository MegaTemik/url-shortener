FROM golang:1.26.1
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o url-shortener ./cmd/url-shortener
CMD ["./url-shortener"]