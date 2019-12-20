package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	slog "github.com/go-eden/slf4go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ms-clovis/snippetbox/pkg/infrastructure"
	"log"
	"net/http"
	"os"
)

func main() {
	// set and parse flags --addr
	addr := flag.String("addr", ":8080", "web server's listening address")
	dsn := flag.String("dsn", "web:password@/snippetbox?parseTime=true", "MySQL data source name")
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

	//s.Router.SetFuncMap(template.FuncMap{"displayDate":handlers.DisplayDate})

	infoLog.Println("Opening database snippetbox on default port")
	s.SetRepo("mysql", *dsn)
	defer s.SnippetRepo.DB.Close()

	// add custom loggers
	//s.ErrorLog = errorLog
	//s.InfoLog = infoLog
	s.Routes()

	infoLog.Println("Starting server on ", *addr)
	err := http.ListenAndServe(*addr, s.Router)
	if err != nil {
		errorLog.Fatal(err)
	}
}
