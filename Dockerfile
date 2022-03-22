FROM golang:1.17-alpine

COPY ./modelo /go/src/app/modelo
WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
COPY server.* ./
COPY .env* ./

RUN go build -o /docker-uliparser

CMD [ "/docker-uliparser" ]
