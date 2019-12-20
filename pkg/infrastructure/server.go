package infrastructure

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	slog "github.com/go-eden/slf4go"
	"github.com/ms-clovis/snippetbox/pkg/handlers"
	"github.com/ms-clovis/snippetbox/pkg/models"
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
	SnippetRepo *mysql.SnippetRepo
	Router      *gin.Engine
	//// logging (for now)
	//ErrorLog *log.Logger
	//InfoLog *log.Logger
	Slog *slog.Logger
}

func NewServer() *Server {
	slog.Debug("Should not see this")

	s := &Server{}
	s.Slog = slog.GetLogger()
	return s
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

}

func (s *Server) HandleHomePage() http.HandlerFunc {

	files := []string{
		"./ui/html/home.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("home.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
		s.logPathAndMethod(r)

		if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
			return
		}
		if r.URL.Path != "/" && r.URL.Path != "/home" {
			s.Slog.Error("Incorrect Path: " + r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			s.Slog.Error("Incorrect Method: " + r.Method)
			http.NotFound(w, r)
			return
		}
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
		tmpl = template.Must(template.New(fileName).Funcs(template.FuncMap{"displayDate": handlers.DisplayDate}).ParseFiles(files...))
		//tmpl = template.Must(template.ParseFiles(files...))
	})
	return tmpl
}

func (s *Server) HandleDisplaySnippet() http.HandlerFunc {

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

		urlStr := r.URL.String()
		s.Slog.Info(urlStr)
		var idStr string = "0"
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
		s.CatchTemplateErrors(tmpl, snippet, w)
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
		if utf8.RuneCountInString(title) > 280 {
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
		http.Redirect(w, req, fmt.Sprintf("/snippet/display/%v", snippet.ID), http.StatusSeeOther)
	}
}

func (s *Server) HandleLatestSnippet() http.HandlerFunc {
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
		s.CatchTemplateErrors(tmpl, snippet, w)
	}
}

func (s *Server) isCorrectHttpMethod(r *http.Request, w http.ResponseWriter, method string) bool {
	if r.Method != method {

		s.Slog.Errorf("Method %v is wrong Http Method", method)
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
	Errors  map[string]string
}

func (s Server) HandleShowSnippetForm(fv FormVals) http.HandlerFunc {
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
