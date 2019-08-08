package main

import (
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe(":1234", http.FileServer(http.Dir("./"))))
}
