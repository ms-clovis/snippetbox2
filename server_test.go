package snippetbox

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ms-clovis/snippetbox/pkg/infrastructure"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"github.com/ms-clovis/snippetbox/pkg/repository/mysql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_HandleHome(t *testing.T) {
	s, mock := setUpServerTesting(t)
	defer s.SnippetRepo.CloseDB()
	snippet := &models.Snippet{
		ID:      1,
		Title:   "Test snippet",
		Content: "I am a test snippet",
		Created: time.Now(),
		Expires: time.Now().Add(time.Hour),
		Author:  "bar@test.com",
	}

	rows := mock.NewRows([]string{"id", "title", "content", "created", "expired", "author"}).
		AddRow(snippet.ID, snippet.Title,
			snippet.Content, snippet.Created, snippet.Expires, snippet.Author)
	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	h := s.HandleHomePage(nil)

	ts := httptest.NewServer(h)
	defer ts.Close()
	req := httptest.NewRequest(http.MethodGet, "/home", nil)

	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Fatalf("expected status code to be 200, but got: %d", resp.Code)
	}
	//fmt.Println("__________________")
	//fmt.Println(resp.Body)

}

func TestServer_WrongMethod(t *testing.T) {
	s, _ := setUpServerTesting(t)
	defer s.SnippetRepo.CloseDB()

	h := s.HandleCreateSnippet()
	ts := httptest.NewServer(h)
	defer ts.Close()

	// body will eventually be snippet values !!!!!

	req := httptest.NewRequest(http.MethodPut, "/snippet/create", nil)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	result := resp.Result()
	fmt.Println(result.Status)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status code to be 405, but got: %d", resp.Code)
	}

}

func setUpServerTesting(t *testing.T) (*infrastructure.Server, sqlmock.Sqlmock) {
	s := infrastructure.NewServer()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	s.SnippetRepo = mysql.NewSnippetRepo(db)
	return s, mock
}

func TestServer_HandleCreateSnippet(t *testing.T) {
	//s := NewServer()
	//	//db, mock, err := sqlmock.New()
	//	//if err != nil {
	//	//	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	//	//}
	//	//
	//	//s.SnippetRepo = mysql.NewSnippetRepo(db)
	s, mock := setUpServerTesting(t)
	defer s.SnippetRepo.CloseDB()

	snippet := &models.Snippet{
		ID:      1,
		Title:   "Test snippet",
		Content: "I am a test snippet",
		Created: time.Now(),
		Expires: time.Now().Add(time.Hour),
	}
	mock.ExpectExec("INSERT ").
		WithArgs(snippet.Title, snippet.Content, snippet.Created, snippet.Expires).
		WillReturnResult(sqlmock.NewResult(0, 1))

	h := s.HandleCreateSnippet()

	ts := httptest.NewServer(h)
	defer ts.Close()

	// body will eventually be snippet values !!!!!

	req := httptest.NewRequest("POST", "/snippet/create", nil)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Fatalf("expected status code to be 200, but got: %d", resp.Code)
	}
	fmt.Println("__________________")
	fmt.Println(resp.Body)
}
