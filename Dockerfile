FROM golang:alpine

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /app

COPY . /app

RUN go get -u github.com/valyala/fasthttp
RUN go get -u github.com/buger/jsonparser

#EXPOSE 80

RUN go build

CMD ./app