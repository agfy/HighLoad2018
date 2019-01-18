package main

import (
	"github.com/valyala/fasthttp"
	"log"
	"runtime"
)

var db = initializeSchema()

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	s := fasthttp.Server{
		Handler: handler,
	}

	err := s.ListenAndServe("127.0.0.1:80")
	if err != nil {
		log.Fatalf("error in ListenAndServe: %s", err)
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("Hello, world!\n")
}