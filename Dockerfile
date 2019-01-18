FROM golang:alpine

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

RUN ["go", "get", "-u", "github.com/valyala/fasthttp"]

# Make port 80 available to the world outside this container
EXPOSE 80

# Run go build when the container launches
RUN ["go", "build"]

CMD ["./app"]