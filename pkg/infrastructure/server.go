package infrastructure

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	slog "github.com/go-eden/slf4go"
	"github.com/golangcollege/sessions"
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
	Session *sessions.Session
	Slog    *slog.Logger
}

func NewServer() *Server {
	slog.Debug("Should not see this")

	s := &Server{}
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

func (s *Server) HandleLoginShowForm(data interface{}) http.HandlerFunc {
	s.Slog.Info("Handle Login Show Form")
	files := []string{
		"./ui/html/login.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("login.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
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

		s.CatchTemplateErrors(tmpl, data, w)

	}
}

func (s *Server) HandleHomePage() http.HandlerFunc {
	s.Slog.Info("Handle Home Page")
	files := []string{
		"./ui/html/home.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("home.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
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
		data := struct {
			Title    string
			Snippets []*models.Snippet
		}{
			Title:    "Home",
			Snippets: snippets,
		}

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

		//w.Header().Set("Content-Type", "application/json")
		// use above for json responses

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
		flash := s.Session.GetString(r, "flash")
		//fmt.Println(flash)
		s.Session.Remove(r, "flash")
		data := struct {
			Message string
			Snippet *models.Snippet
		}{
			Message: flash,
			Snippet: snippet,
		}
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
		fv := FormVals{
			Snippet: snippet,
			Errors:  nil,
		}

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
			fv.Errors = e
			s.HandleShowSnippetForm(fv).ServeHTTP(w, req)

			return
		}

		id, err := s.SnippetRepo.Create(snippet)

		if err != nil {
			log.Fatal(err)
		}
		snippet.ID = int(id)
		s.Session.Put(req, "flash", "Snippet successfully created!")
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
		flash := s.Session.GetString(r, "flash")
		s.Session.Remove(r, "flash")
		//fmt.Println(flash)
		data := struct {
			Message string
			Snippet *models.Snippet
		}{
			Message: flash,
			Snippet: snippet,
		}
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

func (s Server) HandleShowSnippetForm(fv FormVals) http.HandlerFunc {
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
		//data := struct{
		//	Errors map [string]string
		//}{
		//	Errors:errors,
		//}
		s.CatchTemplateErrors(tmpl, fv, w)
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
			data := struct {
				Errors map[string]string
				User   *models.User
			}{
				Errors: errs,
				User:   user,
			}
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

			u := &models.User{

				Name: emailName,

				Active: true,
			}
			u.SetEncryptedPassword(password)
			_, err = s.UserRepo.Create(u)
			if err != nil {
				s.serverError(w, err)
			}
			setSessionIDCookie(w, u.Password)

			//s.HandleHomePage().ServeHTTP(w, r)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// check to see if logged In

		isAuth, err := s.UserRepo.IsAuthenticated(user)
		if err != nil && err != models.ERRNoRecordFound {
			s.serverError(w, err)
			return
		}

		if !isAuth {
			// change error map
			for k := range errs {
				delete(errs, k)
			}
			errs["Password"] = "Your UserName / Password is incorrect"
			user := &models.User{Name: emailName, Password: password}
			data := struct {
				Errors map[string]string
				User   *models.User
			}{
				Errors: errs,
				User:   user,
			}
			w.WriteHeader(http.StatusSeeOther)
			s.HandleLoginShowForm(data).ServeHTTP(w, r)
			return
		}

		// redirect to / with message
		setSessionIDCookie(w, user.Password)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	}

}

func setSessionIDCookie(w http.ResponseWriter, hashedPW string) {
	// need to create and store sessionid
	cookie := http.Cookie{
		Name:    "sessionid",
		Value:   hashedPW,
		Path:    "/",
		Expires: time.Now().Add(time.Hour),
	}
	http.SetCookie(w, &cookie)

}
