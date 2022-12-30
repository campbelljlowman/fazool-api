FROM golang:1.19.3
WORKDIR /usr/src/fazool-api

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o app .

EXPOSE 8080

CMD ["./app"]
