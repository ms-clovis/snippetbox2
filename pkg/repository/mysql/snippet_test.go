package mysql

import (
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"testing"
	"time"
)

func TestSnippetRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sr := NewSnippetRepo(db)

	defer sr.DB.Close()

	snippet := &models.Snippet{

		Title:   "Test snippet",
		Content: "I am a test snippet",
		Created: time.Now(),
		Expires: time.Now().Add(time.Hour),
	}
	//mock.ExpectPrepare("INSERT")
	mock.ExpectExec("INSERT").WithArgs(snippet.Title, snippet.Content, snippet.Created, snippet.Expires).
		WillReturnResult(sqlmock.NewResult(0, 1))

	id, err := sr.Create(snippet)
	if err != nil || int(id) != snippet.ID {
		t.Error(err)
		t.Error("Did not create snippet")
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

func TestSnippetRepo_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sr := NewSnippetRepo(db)
	defer sr.DB.Close()
	mock.ExpectExec("DELETE").
		WillReturnResult(sqlmock.NewResult(0, 1))
	snippet := &models.Snippet{
		ID:      1,
		Title:   "Test snippet",
		Content: "I am a test snippet",
		Created: time.Now(),
		Expires: time.Now().Add(time.Hour),
	}

	wasDeleted, err := sr.Delete(snippet)
	//snippet = nil
	if err != nil || !wasDeleted {
		t.Error("Snippet not deleted from DB")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

func TestSnippetRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sr := NewSnippetRepo(db)
	defer sr.DB.Close()

	snippet := &models.Snippet{
		ID:      1,
		Title:   "Test snippet",
		Content: "I am a test snippet",
		Created: time.Now(),
		Expires: time.Now().Add(time.Hour),
	}

	rows := sqlmock.NewRows([]string{"id", "title", "content", "created", "expires"}).
		AddRow(snippet.ID, snippet.Title, snippet.Content, snippet.Created, snippet.Expires)
	mock.ExpectQuery("SELECT").WithArgs(1).
		WillReturnRows(rows)
	s, err := sr.GetByID(snippet.ID)
	if err != nil {
		t.Error(err)
	}
	if *s != *snippet {
		t.Error("Did not get correct snippet")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
		t.Error("Not all expectations were met")
	}
}
