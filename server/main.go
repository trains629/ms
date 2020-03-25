package main

import (
	"flag"
	"net/http"
)

func main() {
	port := flag.String("port", ":8084", "port")
	flag.Parse()
	http.ListenAndServe(*port, NewPostHandler())
}
