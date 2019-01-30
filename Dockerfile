FROM golang:alpine AS build_base

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /app
COPY . /app

RUN ["go", "get", "-u", "github.com/valyala/fasthttp"]
RUN ["go", "get", "-u", "github.com/buger/jsonparser"]
RUN ["go", "get", "-u", "github.com/lib/pq"]
RUN ["go", "build"]

FROM postgres:alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=0 /app/app /app
#COPY "./data.zip" /tmp/data/
COPY "./wait-for-it.sh" /app

#EXPOSE 80

CMD /docker-entrypoint.sh postgres & ./wait-for-it.sh 0.0.0.0:5432 -- echo "database is up" && ./app