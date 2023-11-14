FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN go build -o /api

EXPOSE 80
EXPOSE 443

CMD [ "/api" ]