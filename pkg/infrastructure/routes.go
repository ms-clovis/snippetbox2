package infrastructure

import (
	"github.com/ms-clovis/snippetbox/pkg/handlers"
	"net/http"
)

func (s Server) Routes() {

	s.Router.Handle("/", handlers.SecureHeaders(s.HandleHomePage()))
	s.Router.HandleFunc("/snippet", handlers.SecureHeaders(s.HandleDisplaySnippet()))
	s.Router.HandleFunc("/snippet/create", handlers.SecureHeaders(s.HandleCreateSnippet()))
	s.Router.Handle("/latest", handlers.SecureHeaders(s.HandleLatestSnippet()))
	// strip prefix LOOKS ONLY for paths that begin with the prefix and then use the FileServer (in this case)
	// handler. The File Server is looking for paths (after the "stripping" of the prefix) and adding them to
	// the directory on the hard drive listed

	s.Router.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static/"))))
}
