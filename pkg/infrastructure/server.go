package infrastructure

import (
	"bytes"
	"database/sql"
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"github.com/ms-clovis/snippetbox/pkg/repository/mysql"
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
)

type Server struct {
	//Repo   *sql.DB
	SnippetRepo *mysql.SnippetRepo
	Router      *http.ServeMux
	//// logging (for now)
	//ErrorLog *log.Logger
	//InfoLog *log.Logger
	Slog *slog.Logger
}

func NewServer() Server {
	slog.Debug("Should not see this")

	s := Server{}
	s.Slog = slog.GetLogger()
	return s
}

func (s *Server) logPathAndMethod(r *http.Request) {
	s.Slog.Info("Path: " + r.URL.Path)
	s.Slog.Info("Method: " + r.Method)
	//s.InfoLog.Println("Path: "+ r.URL.Path)
	//s.InfoLog.Println("Method: "+ r.Method)
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

	var init sync.Once
	var tmpl *template.Template

	s.Slog.Info("Initializing tmpl")
	init.Do(func() {
		files := []string{
			"./ui/html/home.page.html",
			"./ui/html/base.layout.html",
			"./ui/html/footer.partial.html",
		}
		tmpl = template.Must(template.ParseFiles(files...))
	})
	return func(w http.ResponseWriter, r *http.Request) {
		s.logPathAndMethod(r)
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

func (s *Server) HandleDisplaySnippet() http.HandlerFunc {
	var tmpl *template.Template
	var init sync.Once
	files := []string{
		"./ui/html/show.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	// is this better than a map??
	init.Do(func() {
		s.Slog.Info("Parsed Template(s) first time")
		tmpl = template.Must(template.ParseFiles(files...))
	})
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "application/json")
		// use above for json responses

		s.logPathAndMethod(r)

		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id < 1 {
			s.clientError(w, http.StatusNoContent)
			return
		}
		snippet, err := s.SnippetRepo.GetByID(id)
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
		if req.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			//w.WriteHeader(405)
			//http.Error(w, "Method Not Allowed", 405)
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}
		_, err := w.Write([]byte("Create a new snippet..."))
		if err != nil {
			log.Fatal(err)
		}
	}
}
