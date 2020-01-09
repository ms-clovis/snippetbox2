package infrastructure

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	slog "github.com/go-eden/slf4go"
	"github.com/golangcollege/sessions"
	"github.com/justinas/nosurf"
	"github.com/ms-clovis/snippetbox/pkg/handlers/validation"
	"github.com/ms-clovis/snippetbox/pkg/handlers/web"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"github.com/ms-clovis/snippetbox/pkg/repository"
	"github.com/ms-clovis/snippetbox/pkg/repository/mysql"
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type Server struct {
	//Repo   *sql.DB
	HttpServer  *http.Server
	SnippetRepo repository.SnippetRepository
	UserRepo    repository.UserRepository
	Router      *gin.Engine
	//// logging (for now)
	//ErrorLog *log.Logger
	//InfoLog *log.Logger
	Session    *sessions.Session
	SessionMap map[string]*models.User
	Slog       *slog.Logger
}

func checkSessionMapForExpiredSessions() {
	// not ready for this yet
}

func NewServer() *Server {
	//slog.Debug("Should not see this")

	s := &Server{}
	ticker := time.NewTicker(time.Hour)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// check for expired user sessions
				checkSessionMapForExpiredSessions()

			}
		}
	}()
	s.Slog = slog.GetLogger()
	return s
}

func (s *Server) SetHttpServer(server *http.Server) {
	s.HttpServer = server
}

func (s *Server) logPathAndMethod(r *http.Request) {
	s.Slog.Info("Path: " + r.URL.Path)
	s.Slog.Info("Method: " + r.Method)
	//s.Slog.Info("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

}

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (s *Server) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	s.Slog.Error(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (s *Server) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (s *Server) notFound(w http.ResponseWriter) {
	s.clientError(w, http.StatusNotFound)
}

func (s *Server) SetRepo(driverName string, dsnString string) {
	repo, err := sql.Open(driverName, dsnString)
	if err != nil {
		log.Fatal(err)
	}
	if err = repo.Ping(); err != nil {

		log.Fatal(err)
	}

	s.SnippetRepo = mysql.NewSnippetRepo(repo)
	s.UserRepo = mysql.NewUserRepository(repo)

}

func (s *Server) HandleLoginShowForm(data *web.DataVals) http.HandlerFunc {
	s.Slog.Info("Handle Login Show Form")
	files := []string{
		"./ui/html/login.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("login.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
		if data == nil || data.User == nil {
			data = s.getDefaultDataVals(data, r)
		}
		s.logPathAndMethod(r)

		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/display/login" && r.URL.Path != "/user/login" {
			s.Slog.Error("Incorrect Path: " + r.URL.Path)
			http.NotFound(w, r)
			return
		}
		s.RemoveSessionInfo(r, w)

		s.CatchTemplateErrors(tmpl, data, w)

	}
}

func (s *Server) HandleHomePage(data *web.DataVals) http.HandlerFunc {
	s.Slog.Info("Handle Home Page")

	files := []string{
		"./ui/html/home.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("home.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
		if data == nil || data.User.ID == 0 {
			data = s.getDefaultDataVals(data, r)
		}

		s.logPathAndMethod(r)

		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			s.Slog.Error("Wrong method: " + r.Method)
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}
		//if r.URL.Path != "/" && r.URL.Path != "/home" {
		//	s.Slog.Error("Incorrect Path: " + r.URL.Path)
		//	http.NotFound(w, r)
		//	return
		//}

		snippets, err := s.SnippetRepo.Fetch(10)
		if err != nil {
			s.Slog.Error(err)
			http.NotFound(w, r)
			return
		}

		data.Snippets = snippets
		data.Title = "Home"

		s.CatchTemplateErrors(tmpl, data, w)
	}
}

func (s *Server) ParseTemplates(fileName string, files []string) *template.Template {
	var tmpl *template.Template
	var init sync.Once

	// is this better than a map??
	init.Do(func() {
		s.Slog.Info("Parsed Template(s) first time")
		//s.Router.SetFuncMap(template.FuncMap{"displayDate":handlers.DisplayDate})
		tmpl = template.Must(template.New(fileName).Funcs(template.FuncMap{"displayDate": web.DisplayDate}).ParseFiles(files...))
		//tmpl = template.Must(template.ParseFiles(files...))
	})
	return tmpl
}

func (s *Server) HandleDisplaySnippet() http.HandlerFunc {
	s.Slog.Info("Handle Display snippet")
	files := []string{
		"./ui/html/show.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("show.page.html", files)
	return func(w http.ResponseWriter, r *http.Request) {

		s.logPathAndMethod(r)
		//if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
		//	return
		//}

		urlStr := r.URL.String()
		s.Slog.Info(urlStr)
		var idStr = "0"
		urlVals := strings.Split(urlStr, "/")
		if len(urlVals) == 4 {
			idStr = urlVals[3]
		}
		var id int
		if strings.TrimSpace(idStr) == "" { // empty value for ID
			idStr = "0"
		}
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		snippet, err := s.SnippetRepo.GetByID(id)
		if err != nil {
			s.Slog.Error(err)
			if err == models.ERRNoRecordFound {
				s.clientError(w, http.StatusBadRequest)
			} else {
				http.NotFound(w, r)
			}
			return

		}
		var flash = ""
		if s.Session != nil {
			flash = s.Session.GetString(r, "flash")
			//fmt.Println(flash)
			s.Session.Remove(r, "flash")
		}

		data := &web.DataVals{
			Message: flash,
			Snippet: snippet,
			Title:   "Newest Snippet",
		}
		s.getDefaultDataVals(data, r)
		s.CatchTemplateErrors(tmpl, data, w)
	}
}

func (s *Server) CatchTemplateErrors(tmpl *template.Template, data interface{}, w http.ResponseWriter) {
	// in case of templates with errors
	// and you don't want to show partial pages
	// Initialize a new buffer.
	buf := new(bytes.Buffer)
	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our serverError helper and then
	// return.
	err := tmpl.Execute(buf, data)
	//err = tmpl.ExecuteTemplate(w, "show.page.html", snippet)
	if err != nil {
		s.Slog.Error(err)
		s.serverError(w, err)
		return
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		s.Slog.Error(err)
		s.serverError(w, err)
		return
	}
}

func (s *Server) HandleCreateSnippet() http.HandlerFunc {
	s.Slog.Info("Handle Create Snippet")
	return func(w http.ResponseWriter, req *http.Request) {

		s.logPathAndMethod(req)
		if !s.isCorrectHttpMethod(req, w, http.MethodPost) {
			return
		}

		err := req.ParseForm()
		if err != nil {
			s.Slog.Error(err)
			s.serverError(w, err)
		}
		title := req.PostForm.Get("title")
		content := req.PostForm.Get("content")
		expiresDays := req.PostForm.Get("expires")

		if expiresDays == "" {
			expiresDays = "1" // give radio button a default
		}
		intExpiresDays, err := strconv.Atoi(expiresDays)
		if err != nil {
			s.Slog.Error(err)
			s.serverError(w, err)
		}
		e := make(map[string]string)
		snippet := &models.Snippet{

			Title:   title,
			Content: content,
			Created: time.Now(),
			Expires: time.Now().Add(time.Hour * time.Duration(24) * time.Duration(intExpiresDays)),
		}

		//get user from map

		data := s.getDefaultDataVals(nil, req)
		if strings.TrimSpace(title) == "" {
			//e = append(e,"Must have title")
			e["Title"] = "The Snippet must have a title"
		}
		if utf8.RuneCountInString(title) > 50 {
			e["Title"] = "The title can not be more than 100 characters"
		}
		if strings.TrimSpace(content) == "" {
			//e = append(e,"Must have content")
			e["Content"] = "The Snippet must have content"
		}
		if utf8.RuneCountInString(content) > 280 {
			e["Content"] = "The content can not be more than 280 characters"
		}

		if len(e) > 0 {
			data.Errors = e
			data.Title = "Create Snippet"
			data.Snippet = models.NewEmptySnippet()
			data.Snippet.Title = title
			data.Snippet.Content = content
			s.HandleShowSnippetForm(data).ServeHTTP(w, req)

			return
		}

		id, err := s.SnippetRepo.Create(snippet)

		if err != nil {
			log.Fatal(err)
		}
		snippet.ID = int(id)
		if s.Session != nil {
			s.Session.Put(req, "flash", "Snippet successfully created!")
		}

		http.Redirect(w, req, fmt.Sprintf("/snippet/display/%v", snippet.ID), http.StatusSeeOther)
	}
}

func (s *Server) HandleLatestSnippet() http.HandlerFunc {
	s.Slog.Info("Handle Latest Snippet")
	files := []string{
		"./ui/html/show.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("show.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "application/json")
		// use above for json responses

		s.logPathAndMethod(r)
		if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
			return
		}

		snippet, err := s.SnippetRepo.Latest()
		if err != nil {
			s.Slog.Error(err)
			if err == models.ERRNoRecordFound {
				s.clientError(w, http.StatusNoContent)
			} else {
				http.NotFound(w, r)
			}
			return

		}
		flash := ""
		if s.Session != nil {
			flash = s.Session.GetString(r, "flash")
			s.Session.Remove(r, "flash")
		}

		data := &web.DataVals{
			Title:   "Lastest Snippet",
			Message: flash,
			Snippet: snippet,
		}
		data = s.getDefaultDataVals(data, r)
		s.CatchTemplateErrors(tmpl, data, w)
	}
}

func (s *Server) isCorrectHttpMethod(r *http.Request, w http.ResponseWriter, method string) bool {
	if r.Method != method {
		s.Slog.Error(r.Method)
		s.Slog.Errorf("Method %v is wrong Http Method", r.Method)
		w.Header().Set("Allow", method)
		//w.WriteHeader(405)
		//http.Error(w, "Method Not Allowed", 405)
		s.clientError(w, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

type FormVals struct {
	Snippet *models.Snippet
	User    *models.User
	Errors  map[string]string
}

func (s Server) getDefaultDataVals(dv *web.DataVals, r *http.Request) *web.DataVals {
	if dv == nil {
		dv = &web.DataVals{}
	}
	sessionID, err := r.Cookie("sessionid")
	if err != nil {
		sessionID = &http.Cookie{Value: "none"}
	}
	user, ok := s.SessionMap[sessionID.Value]
	if !ok {
		user = &models.User{}
	}

	dv.User = user
	dv.IsAuthenticated = user.Active
	dv.CurrentYear = time.Now().Year()
	dv.CSRFToken = nosurf.Token(r)

	return dv
}

func (s Server) HandleShowSnippetForm(data *web.DataVals) http.HandlerFunc {
	s.Slog.Info("Handle show Snippet form")

	files := []string{
		"./ui/html/create.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("create.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
		s.logPathAndMethod(r)
		//if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
		//	return
		//}
		data = s.getDefaultDataVals(data, r)

		s.CatchTemplateErrors(tmpl, data, w)
	}
}

func (s *Server) HandleLoginRegistration() http.HandlerFunc {
	s.Slog.Info("Handle Login registration")
	return func(w http.ResponseWriter, r *http.Request) {
		s.logPathAndMethod(r)
		s.isCorrectHttpMethod(r, w, http.MethodPost)
		err := r.ParseForm()
		if err != nil {
			slog.Error(err)
			s.serverError(w, err)
		}

		emailName := r.PostFormValue("Name")
		s.Slog.Info(emailName)
		password := r.PostFormValue("Password")
		//s.Slog.Info(password)
		errs := make(map[string]string)
		if validation.IsBlank(emailName) {
			errs["Name"] = models.ERRMustHaveName.Error()
		}
		if !validation.IsValidEmailAddr(emailName) {
			errs["Name"] = models.ERRMustBeValidEmailAddress.Error()
		}
		if validation.IsBlank(password) {
			errs["Password"] = models.ERRMustHavePassword.Error()
		}
		if validation.IsLessThanChars(password, 6) {
			errs["Password"] = fmt.Sprintf("Password Must Have %v Characters", 6)
		}

		if len(errs) > 0 {
			user := &models.User{Name: emailName, Password: password}

			data := s.getDefaultDataVals(nil, r)

			data.Title = "Login - Registration"
			data.Errors = errs
			data.User = user
			w.WriteHeader(http.StatusSeeOther)
			s.HandleLoginShowForm(data).ServeHTTP(w, r)
			return
		}
		user, err := s.UserRepo.GetUser(emailName)
		if err != nil && err != models.ERRNoUserFound {
			s.serverError(w, err)
			return
		}

		if user == nil {
			// create user

			user = s.CreateUser(emailName, password, w)
			s.SessionMap[user.Password] = user

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// check to see if logged In

		isAuth := s.UserRepo.IsAuthenticated(user.Password, password)

		if !isAuth {
			// change error map
			for k := range errs {
				delete(errs, k)
			}
			errs["General"] = "Your UserName / Password is incorrect"

			//sessionID, err := r.Cookie("sessionid")
			//if err != nil {
			//	sessionID = &http.Cookie{Value:"none"}
			//}
			//u,ok := s.SessionMap[sessionID.Value]
			//if ok{
			//	// use user in sessionMap
			//	user = u
			//}
			//otherwise use user from database
			data := s.getDefaultDataVals(nil, r)

			data.User = user
			data.Errors = errs
			data.Title = "Login - Registration"

			w.WriteHeader(http.StatusSeeOther)
			s.HandleLoginShowForm(data).ServeHTTP(w, r)
			return
		}

		// redirect to / with message
		s.setSessionIDCookie(w, user.Password)
		s.SessionMap[user.Password] = user
		//data := &web.DataVals{
		//	User:user,
		//	IsAuthenticated:true,
		//	Message:fmt.Sprintf("Welcome back %v",user.Name),
		//}
		data := s.getDefaultDataVals(nil, r)
		data.User = user
		data.Message = fmt.Sprintf("Welcome back %v", user.Name)
		data.IsAuthenticated = true
		s.HandleHomePage(data).ServeHTTP(w, r)
		//http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	}

}

func (s *Server) CreateUser(emailName string, password string, w http.ResponseWriter) *models.User {
	u := &models.User{

		Name: emailName,

		Active: true,
	}
	u.SetEncryptedPassword(password)
	id, err := s.UserRepo.Create(u)
	if err != nil {
		s.serverError(w, err)
	}
	u.ID = id
	s.setSessionIDCookie(w, u.Password)
	return u
}

func (s *Server) setSessionIDCookie(w http.ResponseWriter, hashedPW string) {
	// need to create and store sessionid
	cookie := http.Cookie{
		Name:    "sessionid",
		Value:   hashedPW,
		Path:    "/",
		Expires: time.Now().Add(time.Hour),
	}
	if hashedPW == "" {
		// delete it
		cookie.MaxAge = 0
		cookie.Expires = time.Now().AddDate(0, -1, 0)
	}
	http.SetCookie(w, &cookie)

}

func (s *Server) LoginForNoSession(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionID, err := r.Cookie("sessionid")

		if sessionID == nil || err != nil {
			http.Redirect(w, r, "/display/login", http.StatusSeeOther)

			return
		} else {
			if _, ok := s.SessionMap[sessionID.Value]; !ok {
				s.RemoveSessionInfo(r, w)
				http.Redirect(w, r, "/display/login", http.StatusSeeOther)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) HandleLogOut() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.RemoveSessionInfo(r, w)
		http.Redirect(w, r, "/display/login", http.StatusSeeOther)
	})
}

func (s *Server) RemoveSessionInfo(r *http.Request, w http.ResponseWriter) {
	sessionID, err := r.Cookie("sessionid")
	if err != nil {
		sessionID = &http.Cookie{Value: "none"}
	}
	delete(s.SessionMap, sessionID.Value)
	s.setSessionIDCookie(w, "") // will delete cookie

}
