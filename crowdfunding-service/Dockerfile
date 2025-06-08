FROM golang:1.24.3

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o crowdfunding-service .

CMD ["go", "run", "."]
