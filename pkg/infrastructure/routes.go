package infrastructure

import (
	"net/http"
)

func (s Server) Routes() {

	s.Router.Handle("/", s.HandleHomePage())
	s.Router.HandleFunc("/snippet", s.HandleDisplaySnippet())
	s.Router.HandleFunc("/snippet/create", s.HandleCreateSnippet())

	// strip prefix LOOKS ONLY for paths that begin with the prefix and then use the FileServer (in this case)
	// handler. The File Server is looking for paths (after the "stripping" of the prefix) and adding them to
	// the directory on the hard drive listed

	s.Router.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static/"))))
}
