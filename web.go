package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World! This is the home page.")
}

func main() {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{Port: 80})
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	server := http.Server{
		Handler: http.HandlerFunc(homeHandler),
	}

	err = server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
