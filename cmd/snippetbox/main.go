package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	slog "github.com/go-eden/slf4go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
	"github.com/ms-clovis/snippetbox/pkg/infrastructure"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// set and parse flags --addr
	addr := flag.String("addr", ":8080", "web server's listening address")
	dsn := flag.String("dsn", "web:password@/snippetbox?parseTime=true", "MySQL data source name")
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")

	flag.Parse()

	slog.SetLevel(slog.InfoLevel)
	slog.Debug("Don't see this")
	slog.Info("DO see this")

	// create "custom" loggers using the standard log package
	infoLog := log.New(os.Stdout, "INFO - ", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR - ", log.Ldate|log.Ltime)

	s := infrastructure.NewServer()
	// choose router type
	s.Router = gin.New()

	httpServer := &http.Server{
		Addr:         *addr,
		Handler:      s.Router,
		TLSConfig:    nil,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	s.SetHttpServer(httpServer)

	// choose session management golangCollege sessions (cookie based)
	// chosen

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	s.Session = session

	//s.Router.SetFuncMap(template.FuncMap{"displayDate":handlers.DisplayDate})

	infoLog.Println("Opening database snippetbox on default port")

	// changed Sever repo to interface (still using pointer to concrete struct instance)
	s.SetRepo("mysql", *dsn)
	defer s.SnippetRepo.CloseDB()

	// add custom loggers
	//s.ErrorLog = errorLog
	//s.InfoLog = infoLog
	s.Routes()

	infoLog.Println("Starting server on ", *addr)
	//err := http.ListenAndServe(*addr, s.Router)
	err := s.HttpServer.ListenAndServe()
	if err != nil {
		errorLog.Fatal(err)
	}
}
