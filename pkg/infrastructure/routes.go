package infrastructure

import (
	"github.com/gin-gonic/gin"
	"github.com/ms-clovis/snippetbox/pkg/handlers"
	"net/http"
)

func (s Server) Routes() {

	s.Router.Handle(http.MethodGet, "/", handlers.RecoverPanic(handlers.SecureHeaders(s.HandleHomePage())))
	s.Router.Handle(http.MethodGet, "/snippet", handlers.RecoverPanic(handlers.SecureHeaders(s.HandleDisplaySnippet())))
	s.Router.Handle(http.MethodPost, "/snippet/create", handlers.RecoverPanic(handlers.SecureHeaders(s.HandleCreateSnippet())))
	s.Router.Handle(http.MethodGet, "/latest", handlers.RecoverPanic(handlers.SecureHeaders(s.HandleLatestSnippet())))
	// strip prefix LOOKS ONLY for paths that begin with the prefix and then use the FileServer (in this case)
	// handler. The File Server is looking for paths (after the "stripping" of the prefix) and adding them to
	// the directory on the hard drive listed

	s.Router.Handle(http.MethodGet, "/static/", gin.WrapH(http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static/")))))
}
