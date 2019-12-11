package main

import (
	"github.com/ms-clovis/snippetbox/pkg/infrastructure"
	"log"
	"net/http"
)

func main() {
	s := infrastructure.Server{}
	s.Router = http.NewServeMux()
	s.Routes()
	err := http.ListenAndServe(":8080", s.Router)
	if err != nil {
		log.Fatal(err)
	}
}
