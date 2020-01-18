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
)

type Server struct {
	//Repo   *sql.DB
	HttpServer  *http.Server
	SnippetRepo repository.SnippetRepository
	UserRepo    repository.UserRepository
	FriendRepo  repository.FriendsRepository
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

// creates db connection, tests it, and uses it to create server repos
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
	s.FriendRepo = mysql.NewFriendsRepository(repo)

}

func (s *Server) HandleLoginShowForm(data *web.DataVals) http.HandlerFunc {

	files := []string{
		"./ui/html/login.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("login.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
		s.Slog.Info("Handle Login Show Form")
		s.logPathAndMethod(r)
		if data == nil {
			data = s.getDefaultDataVals(data, r)
		}
		data.CSRFToken = nosurf.Token(r) // need all times for the login
		data.CurrentLink = "LR"

		s.logPathAndMethod(r)
		if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}
		//if r.Method != http.MethodGet && r.Method != http.MethodPost {
		//	s.clientError(w, http.StatusMethodNotAllowed)
		//	return
		//}
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
		if data == nil || !data.IsAuthenticated {
			data = s.getDefaultDataVals(data, r)
		}
		data.CSRFToken = nosurf.Token(r)
		data.CurrentLink = "H"
		s.logPathAndMethod(r)

		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			s.Slog.Error("Wrong method: " + r.Method)
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}

		snippets, err := s.SnippetRepo.Fetch(data.User, 10)
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
		data := s.getDefaultDataVals(nil, r)
		data.CurrentLink = "LA"

		snippet, err := s.SnippetRepo.GetByID(data.User, id)
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

		data = &web.DataVals{
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

func (s *Server) HandleChangePassword() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		s.Slog.Info("Handle Change Password")
		s.logPathAndMethod(r)
		if !s.isCorrectHttpMethod(r, w, http.MethodPost) {
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}
		err := r.ParseForm()
		if err != nil {
			slog.Error(err)
			s.serverError(w, err)
		}
		user := s.getSessionUser(r)
		data := s.getDefaultDataVals(nil, r)
		data.CurrentLink = "CP"

		oldPass := r.PostFormValue("extPass")
		newPass := r.PostFormValue("newPass")
		newPassMatch := r.PostFormValue("newPassMatch")

		errs := make(map[string]string)
		// check for input errors

		if !validation.IsAuthenticated(user.Password, oldPass) {
			errs["ExtPass"] = "Problem with existing password"
		}
		if validation.IsLessThanChars(newPass, 6) {
			errs["NewPass"] = "New Password must have a least 6 characters"
		}
		if oldPass == newPass {
			errs["NewPass"] = "New Password can not be the same as the old password"
		}
		if newPass != newPassMatch {
			errs["NewPassMatch"] = "New Passwords do not match"
		}

		// problem with passwords
		if len(errs) > 0 {
			data.Errors = errs
			r.Method = http.MethodGet
			s.HandleChangePasswordForm(data).ServeHTTP(w, r)
			return
		}
		// else change the password
		user.SetEncryptedPassword(newPass)
		worked, err := s.UserRepo.Update(user)
		if err != nil || !worked {
			s.serverError(w, err)
			return
		}
		data.Message = "Password successfully changed"
		data.User = user
		s.HandleHomePage(data).ServeHTTP(w, r)
		return

	}

}

func (s *Server) HandleChangePasswordForm(data *web.DataVals) http.HandlerFunc {
	s.Slog.Info("Handle Change Password form (display)")

	files := []string{
		"./ui/html/password.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("password.page.html", files)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logPathAndMethod(r)
		if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
			return
		}
		if data == nil || !data.IsAuthenticated {
			data = s.getDefaultDataVals(data, r)
		}
		data.CurrentLink = "CP"
		s.CatchTemplateErrors(tmpl, data, w)
	}
}

func (s *Server) HandleCreateSnippet() http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		s.Slog.Info("Handle Create Snippet")
		s.logPathAndMethod(req)
		if !s.isCorrectHttpMethod(req, w, http.MethodPost) {
			s.clientError(w, http.StatusMethodNotAllowed)
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
		idStr := req.PostFormValue("ID")

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
		data.CurrentLink = "CS"
		if validation.IsBlank(title) {
			//e = append(e,"Must have title")
			e["Title"] = "The Snippet must have a title"
		}
		if validation.IsMoreThanChars(title, 50) {
			e["Title"] = "The title can not be more than 100 characters"
		}
		if validation.IsBlank(content) {
			//e = append(e,"Must have content")
			e["Content"] = "The Snippet must have content"
		}
		if validation.IsMoreThanChars(content, 280) {
			e["Content"] = "The content can not be more than 280 characters"
		}

		if len(e) > 0 {
			data.Errors = e
			data.Title = "Create Snippet"
			data.Snippet = models.NewEmptySnippet()
			data.Snippet.Title = title
			data.Snippet.Content = content
			data.ExpiresDays = expiresDays
			req.Method = http.MethodGet
			s.HandleShowSnippetForm(data).ServeHTTP(w, req)

			return
		}
		var id int64
		if validation.IsBlank(idStr) || idStr == "0" {
			id, err = s.SnippetRepo.Create(data.User, snippet)
			if err != nil {
				log.Fatal(err)
			}
			snippet.ID = int(id)
		} else {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				s.serverError(w, err)
			}
			snippet.ID = id
			worked, err := s.SnippetRepo.Update(data.User, snippet)
			if !worked || err != nil {
				s.serverError(w, err)
			}
		}

		if s.Session != nil {
			if validation.IsBlank(idStr) || idStr == "0" {
				s.Session.Put(req, "flash", "Snippet successfully created!")
			} else {
				s.Session.Put(req, "flash", "Snippet successfully updated!")
			}
		}

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
		s.Slog.Info("Handle Latest Snippet")

		s.logPathAndMethod(r)
		if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}
		data := s.getDefaultDataVals(nil, r)
		data.CurrentLink = "LA"
		snippet, err := s.SnippetRepo.Latest(data.User)
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

		//data := &web.DataVals{
		//	Title:   "Lastest Snippet",
		//	Message: flash,
		//	Snippet: snippet,
		//}
		data.Title = "Latest Snippet"
		data.Message = flash
		data.Snippet = snippet

		s.CatchTemplateErrors(tmpl, data, w)
	}
}

func (s *Server) isCorrectHttpMethod(r *http.Request, w http.ResponseWriter, correctMethod string) bool {
	if r.Method != correctMethod {
		//s.Slog.Error(r.Method)
		s.Slog.Errorf("Method %v is wrong Http Method,should be %v", r.Method, correctMethod)
		w.Header().Set("Allow", correctMethod)
		//w.WriteHeader(405)
		//http.Error(w, "Method Not Allowed", 405)
		//s.clientError(w, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func (s Server) getDefaultDataVals(dv *web.DataVals, r *http.Request) *web.DataVals {
	if dv == nil {
		dv = &web.DataVals{}
	}
	user := s.getSessionUser(r)

	dv.User = user
	dv.IsAuthenticated = user.Active
	dv.CurrentYear = time.Now().Year()
	dv.CSRFToken = nosurf.Token(r)
	if dv.Snippet != nil && dv.Snippet.Created.Unix() != -62135596800 {
		dv.ExpiresDays = GetExpiresDays(dv.Snippet)
	}
	return dv
}

func GetExpiresDays(s *models.Snippet) string {
	timeDiff := s.Expires.Sub(s.Created).Round(time.Hour)
	days := timeDiff.Hours() / 24
	return fmt.Sprintf("%v", days)
}

func (s Server) getSessionUser(r *http.Request) *models.User {
	sessionID, err := r.Cookie("sessionid")
	if err != nil {
		sessionID = &http.Cookie{Value: "none"}
	}
	user, ok := s.SessionMap[sessionID.Value]
	if !ok {
		user = &models.User{}
	}
	return user
}

func (s Server) HandleShowSnippetForm(data *web.DataVals) http.HandlerFunc {

	files := []string{
		"./ui/html/create.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("create.page.html", files)

	return func(w http.ResponseWriter, r *http.Request) {
		s.Slog.Info("Handle show Snippet form")
		s.logPathAndMethod(r)
		if !s.isCorrectHttpMethod(r, w, http.MethodGet) {
			s.clientError(w, http.StatusMethodNotAllowed)
			return
		}
		err := r.ParseForm()
		if err != nil {
			s.serverError(w, err)
		}
		idStr := r.FormValue("ID")

		data = s.getDefaultDataVals(data, r)
		data.CurrentLink = "CS"
		if idStr != "" && idStr != "0" {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				s.clientError(w, http.StatusNoContent)
			}
			snippet, err := s.SnippetRepo.GetByID(data.User, id)
			if err != nil {
				s.serverError(w, err)
			}
			data.Snippet = snippet
			data.ExpiresDays = GetExpiresDays(snippet)
		}
		s.CatchTemplateErrors(tmpl, data, w)
	}
}

func (s *Server) HandleLoginRegistration() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		s.Slog.Info("Handle Login registration")
		s.logPathAndMethod(r)
		if !s.isCorrectHttpMethod(r, w, http.MethodPost) {
			s.clientError(w, http.StatusMethodNotAllowed)
		}
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
			data.CurrentLink = "LR"

			data.Title = "Login - Registration"
			data.Errors = errs
			data.User = user
			w.WriteHeader(http.StatusSeeOther)
			r.Method = http.MethodGet
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
			//friends,err := s.FriendRepo.FindFriends(user)
			//if err !=nil{
			//	s.serverError(w,err)
			//}
			//user.SetFriendsMap(friends)
			s.SessionMap[user.Password] = user

			r.Method = http.MethodGet
			data := s.getDefaultDataVals(nil, r)
			data.User = user
			data.IsAuthenticated = true
			data.CSRFToken = nosurf.Token(r)
			data.Message = ""
			s.HandleHomePage(data).ServeHTTP(w, r)
			//http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// check to see if logged In

		isAuth := validation.IsAuthenticated(user.Password, password)

		if !isAuth {
			// change error map
			for k := range errs {
				delete(errs, k)
			}
			errs["General"] = "Your UserName / Password is incorrect"

			//otherwise use user from database
			data := s.getDefaultDataVals(nil, r)

			data.User = user
			data.Errors = errs
			data.Title = "Login - Registration"
			r.Method = http.MethodGet
			w.WriteHeader(http.StatusSeeOther)
			s.HandleLoginShowForm(data).ServeHTTP(w, r)
			return
		}

		// redirect to / with message
		s.setSessionIDCookie(w, user.Password)
		s.SessionMap[user.Password] = user

		data := s.getDefaultDataVals(nil, r)
		//friends,err := s.FriendRepo.FindFriends(user)
		//if err !=nil{
		//	s.serverError(w,err)
		//}
		//user.SetFriendsMap(friends)
		data.User = user

		data.Message = fmt.Sprintf("Hello %v", user.Name)

		data.IsAuthenticated = true
		r.Method = http.MethodGet

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
				return
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

func (s Server) HandleFriends() http.HandlerFunc {
	files := []string{
		"./ui/html/friend.page.html",
		"./ui/html/base.layout.html",
		"./ui/html/footer.partial.html",
	}
	tmpl := s.ParseTemplates("friend.page.html", files)
	return func(w http.ResponseWriter, r *http.Request) {
		s.Slog.Info("handle friends")
		switch r.Method {
		case http.MethodGet:
			s.handleShowFriends(w, r, tmpl)
		case http.MethodPost:
			s.handleMakeFriends(w, r, tmpl)
		default:
			s.clientError(w, http.StatusMethodNotAllowed)

		}
	}
}

func (s *Server) handleMakeFriends(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	slog.Info("making friends")
	s.logPathAndMethod(r)
	data := s.getDefaultDataVals(nil, r)
	err := r.ParseForm()
	if err != nil {
		s.serverError(w, err)
		return
	}
	idStr := r.PostFormValue("ID")
	if validation.IsBlank(idStr) {
		s.clientError(w, http.StatusBadRequest)
		return

	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.serverError(w, err)
		return
	}
	friendly, err := s.UserRepo.GetUserByID(id)
	if friendly == nil {
		s.clientError(w, http.StatusBadRequest)
		return
	}
	if data.User.IsFriend(friendly) {
		// unfriend
		unfriend, err := s.FriendRepo.UnFriend(data.User, friendly)
		if err != nil {
			s.serverError(w, err)
			return
		}
		if !unfriend {
			s.clientError(w, http.StatusBadRequest)
			return
		}

	} else {
		// friend
		friended, err := s.FriendRepo.SetFriend(data.User, friendly)
		if err != nil {
			s.serverError(w, err)
			return
		}
		if !friended {
			s.clientError(w, http.StatusBadRequest)
			return
		}
	}

	s.handleShowFriends(w, r, tmpl)

}

func (s *Server) handleShowFriends(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	slog.Info("showing friends")
	s.logPathAndMethod(r)
	err := r.ParseForm()
	if err != nil {
		slog.Error(err)
	}
	urlStr := r.URL.String()
	s.Slog.Info(urlStr)
	var nameStr = ""
	urlVals := strings.Split(urlStr, "/")
	if len(urlVals) == 4 {
		nameStr = urlVals[3]
	}

	//fmt.Println(id)
	data := s.getDefaultDataVals(nil, r)
	data.CurrentLink = "FR"

	friends, err := s.FriendRepo.FindFriends(data.User)
	if err != nil {
		s.serverError(w, err)
	}
	data.User.SetFriendsMap(friends)

	if nameStr == "" || nameStr == "all" {
		others, err := s.UserRepo.GetUsers(data.User)
		if err != nil {
			s.serverError(w, err)
		}
		data.Users = others
	} else {
		other, err := s.UserRepo.GetUser(nameStr)
		if err != nil {
			s.serverError(w, err)
		}
		data.Users = []*models.User{other}
	}

	s.CatchTemplateErrors(tmpl, data, w)

}
