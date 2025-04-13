FROM golang:1.24.1

# Установка зависимостей для CGO и SQLite
RUN apk add --no-cache gcc musl-dev sqlite

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 4000

CMD ["./main"]