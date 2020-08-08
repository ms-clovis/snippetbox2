package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/set", setContext)
	mux.HandleFunc("/get", getContext)
	http.ListenAndServe(":8080", mux)
}

func getContext(w http.ResponseWriter, r *http.Request) {
	var name string
	nameInterface := r.Context().Value("user")
	if nameInterface != nil {
		name = nameInterface.(string)
	}
	_, err := w.Write([]byte(name))
	if err != nil {
		log.Fatal(err)
	}

}

func setContext(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	log.Println(name)
	ctx := context.WithValue(r.Context(), "user", name)
	r.WithContext(ctx)
}
