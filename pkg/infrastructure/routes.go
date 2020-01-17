package infrastructure

import (
	"github.com/gin-gonic/gin"
	"github.com/justinas/alice"
	"github.com/ms-clovis/snippetbox/pkg/handlers/web"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"net/http"
)

func (s Server) Routes() {
	// to use alice must be HANDLERS , not HANDLDERFUNCS, see recoverPanic
	sessionMiddleWare := alice.New(s.Session.Enable, web.RecoverPanic, web.SecureHeaders) //, web.CSRFTokenSetter
	loginRedirectIncl := sessionMiddleWare.Append(s.LoginForNoSession)

	data := &web.DataVals{User: &models.User{}}

	s.Router.Handle(http.MethodGet, "/user/friend/:id", gin.WrapH(loginRedirectIncl.Then(s.HandleFriends())))
	s.Router.Handle(http.MethodPost, "/user/friend", gin.WrapH(loginRedirectIncl.Then(s.HandleFriends())))

	s.Router.Handle(http.MethodPost, "/modify/snippet", gin.WrapH(loginRedirectIncl.Then(s.HandleShowSnippetForm(nil))))
	s.Router.Handle(http.MethodGet, "/modify/snippet", gin.WrapH(loginRedirectIncl.Then(s.HandleShowSnippetForm(nil))))
	s.Router.Handle(http.MethodPost, "/change/password", gin.WrapH(loginRedirectIncl.Then(s.HandleChangePassword())))

	s.Router.Handle(http.MethodGet, "/display/password", gin.WrapH(loginRedirectIncl.Then(s.HandleChangePasswordForm(data))))

	s.Router.Handle(http.MethodGet, "/user/logout", gin.WrapH(loginRedirectIncl.Then(s.HandleLogOut())))
	s.Router.Handle(http.MethodPost, "/user/login", gin.WrapH(sessionMiddleWare.Then(s.HandleLoginRegistration())))

	s.Router.Handle(http.MethodGet, "/display/login", gin.WrapH(sessionMiddleWare.Then(s.HandleLoginShowForm(nil))))

	s.Router.Handle(http.MethodGet, "/", gin.WrapH(loginRedirectIncl.Then(s.HandleHomePage(nil))))
	s.Router.Handle(http.MethodGet, "/snippet/display/:id", gin.WrapH(loginRedirectIncl.Then(s.HandleDisplaySnippet())))
	s.Router.Handle(http.MethodPost, "/snippet/create", gin.WrapH(loginRedirectIncl.Then(s.HandleCreateSnippet())))
	s.Router.Handle(http.MethodGet, "/snippet/create", gin.WrapH(loginRedirectIncl.Then(s.HandleShowSnippetForm(&web.DataVals{Title: "Create Snippet", Snippet: models.NewEmptySnippet()}))))
	s.Router.Handle(http.MethodGet, "/latest", gin.WrapH(loginRedirectIncl.Then(s.HandleLatestSnippet())))

	// strip prefix LOOKS ONLY for paths that begin with the prefix and then use the FileServer (in this case)
	// handler. The File Server is looking for paths (after the "stripping" of the prefix) and adding them to
	// the directory on the hard drive listed

	//s.Router.Handle(http.MethodGet, "/static/", gin.WrapH(http.StripPrefix("/static/", http.FileServer(http.Dir("./ui/static/")))))
	s.Router.Static("/static/", "./ui/static")
}
