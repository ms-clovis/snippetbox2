package mysql

import (
	"database/sql"
	slog "github.com/go-eden/slf4go"
	"github.com/ms-clovis/snippetbox/pkg/models"
)

type SnippetRepo struct {
	DB *sql.DB
}

func NewSnippetRepo(db *sql.DB) *SnippetRepo {
	//fetchPS,err := db.Prepare()
	return &SnippetRepo{DB: db}
}

var FETCHSQL = "SELECT s.id, s.title, s.content, s.created, s.expires, u.name " +
	"FROM snippets s INNER JOIN users u ON s.author = u.id " +
	" LEFT JOIN friends f ON s.Author = f.watched " +
	" WHERE s.expires > UTC_TIMESTAMP() "

//noinspection ALL
func (sr *SnippetRepo) fetch(query string, user *models.User, arg int) ([]*models.Snippet, error) {
	//var ret []models.Snippet
	//var snip models.Snippet
	rows, err := sr.DB.Query(query, user.ID, user.ID, arg)

	defer rows.Close()
	if err != nil {
		slog.Error(err)

		return nil, err
	}
	ret := make([]*models.Snippet, 0)

	for rows.Next() {
		s := &models.Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires, &s.Author)

		ret = append(ret, s)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

//Convenience method to get all Snippets

func (sr *SnippetRepo) FetchAll(user *models.User) ([]*models.Snippet, error) {
	return sr.Fetch(user, 0)
}

// Will retrieve (by limit) up to a certain number of Snippets
// if the number to fetch is 0 it will retrieve all (no limit)

func (sr *SnippetRepo) Fetch(user *models.User, numberToFetch int) ([]*models.Snippet, error) {

	query := FETCHSQL +
		"AND s.author = ? OR s.author in " +
		"(SELECT watched FROM friends WHERE watcher = ? )" +
		" ORDER BY created DESC "

	if numberToFetch > 0 {
		query += " limit ?"

	}
	return sr.fetch(query, user, numberToFetch)

}

func (sr *SnippetRepo) Delete(user *models.User, m *models.Snippet) (bool, error) {
	del := "DELETE FROM snippetbox.snippets WHERE id = ? AND author = ?"
	result, err := sr.DB.Exec(del, m.ID, user.ID)
	if err != nil {
		return false, err
	}

	aff, err := result.RowsAffected()
	return aff == 1, err
}

func (sr *SnippetRepo) GetByID(user *models.User, ID int) (*models.Snippet, error) {
	query := FETCHSQL +
		"AND (s.author = ? OR s.author in (SELECT watched FROM friends WHERE watched = ? ) ) AND s.id = ?  LIMIT 1 "
	snippets, err := sr.fetch(query, user, ID)
	if err != nil {

		return nil, err
	}
	if len(snippets) == 0 {

		return nil, models.ERRNoRecordFound
	}
	return snippets[0], nil

}

func (sr *SnippetRepo) Create(user *models.User, m *models.Snippet) (int64, error) {
	insert := "INSERT INTO snippets (title,content,created,expires,author) VALUES" +
		"( ? , ? , ? , ?,?)"
	result, err := sr.DB.Exec(insert, m.Title, m.Content, m.Created, m.Expires, user.ID)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	m.ID = int(id) // fix this
	return id, err
}

func (sr *SnippetRepo) Update(user *models.User, m *models.Snippet) (bool, error) {
	upd := "UPDATE snippets " +
		"SET title = ? , content = ?, " +
		"created = ?, expires = ? " +
		"WHERE id = ? AND author = ?"
	result, err := sr.DB.Exec(upd, m.Title, m.Content, m.Created, m.Expires, m.ID, user.ID)
	if err != nil {
		return false, err
	}
	aff, err := result.RowsAffected()
	return aff == 1, err

}

func (sr *SnippetRepo) Latest(user *models.User) (*models.Snippet, error) {
	snippets, err := sr.Fetch(user, 1)
	if err != nil {
		return nil, err
	}
	if len(snippets) == 0 {
		return nil, models.ERRNoRecordFound
	}
	return snippets[0], nil
}

func (sr *SnippetRepo) CloseDB() {
	sr.DB.Close()
}
