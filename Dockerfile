FROM golang:1.21

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN apt-get update

# Install PostgreSQL client
RUN apt-get -y install postgresql-client

# Make wait-for-postgres.sh executable if needed
COPY wait-for-postgres.sh /wait-for-postgres.sh
RUN chmod +x /wait-for-postgres.sh

# build go app
RUN go mod download
RUN go build -o diplom ./cmd/main.go

CMD ["./diplom"]