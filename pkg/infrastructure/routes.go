package infrastructure

import (
	"github.com/gin-gonic/gin"
	"github.com/ms-clovis/snippetbox/pkg/handlers/web"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"net/http"
	//"github.com/justinas/alice"
)

func (s Server) Routes() {

	//sessionMiddleWare := alice.New(s.Session.Enable)

	s.Router.Handle(http.MethodGet, "/", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleHomePage())))))
	s.Router.Handle(http.MethodGet, "/snippet/display/:id", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleDisplaySnippet())))))
	s.Router.Handle(http.MethodPost, "/snippet/create", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleCreateSnippet())))))
	s.Router.Handle(http.MethodGet, "/snippet/create", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleShowSnippetForm(FormVals{Snippet: models.NewEmptySnippet()}))))))
	s.Router.Handle(http.MethodGet, "/latest", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleLatestSnippet())))))

	// strip prefix LOOKS ONLY for paths that begin with the prefix and then use the FileServer (in this case)
	// handler. The File Server is looking for paths (after the "stripping" of the prefix) and adding them to
	// the directory on the hard drive listed

	s.Router.Handle(http.MethodGet, "/static/", gin.WrapH(http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static/")))))
}
