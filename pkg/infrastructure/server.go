package infrastructure

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	Repo   *sql.DB
	Router *http.ServeMux
}

func (s Server) handleHomePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Home Page"))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s Server) handleDisplaySnippet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "application/json")
		// use above for json responses

		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id < 1 {
			http.NotFound(w, r)
			return
		}
		_, err = fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s Server) handleCreateSnippet() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			//w.WriteHeader(405)
			http.Error(w, "Method Not Allowed", 405)
			return
		}
		_, err := w.Write([]byte("Create a new snippet..."))
		if err != nil {
			log.Fatal(err)
		}
	}
}
