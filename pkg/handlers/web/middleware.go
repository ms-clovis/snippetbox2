package web

import (
	"fmt"
	slog "github.com/go-eden/slf4go"
	"net/http"
	"runtime/debug"
)

func LoginForNoSession(next http.Handler) http.HandlerFunc {
	//var init sync.Once
	//
	//init.Do(func(){
	//
	//})
	return func(w http.ResponseWriter, r *http.Request) {

		sessionID, err := r.Cookie("sessionid")

		if sessionID == nil || err != nil {
			http.Redirect(w, r, "/display/login", http.StatusSeeOther)

			return
		}
		next.ServeHTTP(w, r)
	}
}

func SecureHeaders(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//slog.Info("Setting secure headers")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		if next != nil {
			next.ServeHTTP(w, r)
		}
	}
}

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//w := ctx.Writer
		//r := ctx.Request
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				slog.Error("Panic error ..")
				trace := fmt.Sprintf("%s\n%s", err, debug.Stack())
				slog.Error(trace)

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
