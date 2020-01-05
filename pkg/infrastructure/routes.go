package infrastructure

import (
	"github.com/gin-gonic/gin"
	"github.com/ms-clovis/snippetbox/pkg/handlers/web"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"net/http"
	//"github.com/justinas/alice"
)

func (s Server) Routes() {
	// to use alice must be HANDLERS , not HANDLDERFUNCS, see recoverPanic
	//sessionMiddleWare := alice.New(s.Session.Enable,web.RecoverPanic)
	//data := struct {
	//	User   models.User
	//	Errors map[string]string
	//}{User: models.User{},
	//	Errors: nil,
	//}

	data := &web.DataVals{User: &models.User{}}
	s.Router.Handle(http.MethodGet, "/user/logout", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleLogOut())))))
	s.Router.Handle(http.MethodPost, "/user/login", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleLoginRegistration())))))

	s.Router.Handle(http.MethodGet, "/display/login", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.HandleLoginShowForm(data))))))

	s.Router.Handle(http.MethodGet, "/", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.LoginForNoSession(s.HandleHomePage(data)))))))
	s.Router.Handle(http.MethodGet, "/snippet/display/:id", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.LoginForNoSession(s.HandleDisplaySnippet()))))))
	s.Router.Handle(http.MethodPost, "/snippet/create", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.LoginForNoSession(s.HandleCreateSnippet()))))))
	s.Router.Handle(http.MethodGet, "/snippet/create", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.LoginForNoSession(s.HandleShowSnippetForm(&web.DataVals{Title: "Create Snippet", Snippet: models.NewEmptySnippet()})))))))
	s.Router.Handle(http.MethodGet, "/latest", gin.WrapH(s.Session.Enable(web.RecoverPanic(web.SecureHeaders(s.LoginForNoSession(s.HandleLatestSnippet()))))))

	// strip prefix LOOKS ONLY for paths that begin with the prefix and then use the FileServer (in this case)
	// handler. The File Server is looking for paths (after the "stripping" of the prefix) and adding them to
	// the directory on the hard drive listed

	//s.Router.Handle(http.MethodGet, "/static/", gin.WrapH(http.StripPrefix("/static/", http.FileServer(http.Dir("./ui/static/")))))
	s.Router.Static("/static/", "./ui/static")
}
