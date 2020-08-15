package snippetbox2

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ms-clovis/snippetbox2/pkg/infrastructure"
	"github.com/ms-clovis/snippetbox2/pkg/models"
	"github.com/ms-clovis/snippetbox2/pkg/repository/mock"
	"github.com/ms-clovis/snippetbox2/pkg/repository/mysql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestServer_HandleDisplaySnippet(t *testing.T) {
	s := infrastructure.NewServer()
	s.SnippetRepo = &mock.MockSnippetRepository{DB: nil}
	defer s.SnippetRepo.CloseDB()

	//s.Session = &sessions.Session{}
	h := s.HandleDisplaySnippet()
	ts := httptest.NewServer(h)
	defer ts.Close()

	tests := []struct {
		Name               string
		URL                string
		WantedResponseCode int
		WantedBody         []byte
	}{
		{Name: "Valid ID", URL: "/snippet/display/1", WantedResponseCode: http.StatusOK, WantedBody: []byte(mock.FakeSnippet.Content)},
		{"Alpha ID", "/snippet/display/1A", http.StatusBadRequest, nil},
		{"Float ID", "/snippet/display/1.23", http.StatusBadRequest, nil},
		{"Empty ID", "/snippet/display/", http.StatusBadRequest, nil},
		{"Trailing Slash with good ID", "/snippet/display/1/", http.StatusBadRequest, nil},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, test.URL, nil)

			resp := httptest.NewRecorder()

			h.ServeHTTP(resp, req)
			if resp.Code != test.WantedResponseCode {
				t.Errorf("Invalid test: %v\nWanted Response: %v\nActual Response: %v\nBody: %v\n",
					test.Name, test.WantedResponseCode, resp.Code, resp.Body)
			}
		})
	}

}

func TestServer_HandleHome(t *testing.T) {
	s, m := setUpServerTesting(t)
	defer s.SnippetRepo.CloseDB()
	snippet := &models.Snippet{
		ID:      1,
		Title:   "Test snippet",
		Content: "I am a test snippet",
		Created: time.Now(),
		Expires: time.Now().Add(time.Hour),
		Author:  "bar@test.com",
	}

	rows := m.NewRows([]string{"id", "title", "content", "created", "expired", "author"}).
		AddRow(snippet.ID, snippet.Title,
			snippet.Content, snippet.Created, snippet.Expires, snippet.Author)
	m.ExpectQuery("SELECT").
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
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	s.SnippetRepo = mysql.NewSnippetRepo(db)
	return s, m
}

func TestServer_HandleCreateSnippet(t *testing.T) {
	//s := NewServer()
	//	//db, mock, err := sqlmock.New()
	//	//if err != nil {
	//	//	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	//	//}
	//	//
	//	//s.SnippetRepo = mysql.NewSnippetRepo(db)
	s, _ := setUpServerTesting(t)
	s.SnippetRepo = &mock.MockSnippetRepository{}
	defer s.SnippetRepo.CloseDB()

	snippet := mock.FakeSnippet
	//m.ExpectExec("INSERT ").
	//	WithArgs(snippet.Title, snippet.Content, snippet.Created, snippet.Expires).
	//	WillReturnResult(sqlmock.NewResult(0, 1))
	h := s.HandleCreateSnippet()

	ts := httptest.NewServer(h)
	defer ts.Close()

	// body will eventually be snippet values !!!!!
	form := url.Values{}
	form.Add("title", snippet.Title)
	form.Add("content", snippet.Content)
	form.Add("expires", "1")
	req := httptest.NewRequest("POST", "/snippet/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	if resp.Code != 303 {
		t.Fatalf("expected status code to be 303, but got: %d", resp.Code)
	}
	fmt.Println("__________________")
	fmt.Println(resp.Body)
}
