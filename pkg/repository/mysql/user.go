package mysql

import (
	"database/sql"
	"errors"
	slog "github.com/go-eden/slf4go"
	"github.com/ms-clovis/snippetbox/pkg/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur *UserRepository) CloseDB() {
	ur.DB.Close()
}

func (ur *UserRepository) Create(u *models.User) (int64, error) {
	insert := "INSERT INTO users( name,password,active)VALUES" +
		"(?,?,?)"

	result, err := ur.DB.Exec(insert, u.Name, u.Password, u.Active)
	if err != nil {
		slog.Error(err)
		return 0, err
	}

	id, err := result.LastInsertId()
	u.ID = id
	return id, err
}

func (ur *UserRepository) fetchByID(query string, id int) (*models.User, error) {
	rows, err := ur.DB.Query(query, id)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	user := &models.User{}
	if rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Password, &user.Active)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("No Matching User")
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *UserRepository) fetch(query string, arg string) (*models.User, error) {
	rows, err := ur.DB.Query(query, arg)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	user := &models.User{}
	if rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Password, &user.Active)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, models.ERRNoUserFound
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *UserRepository) GetUser(name string) (*models.User, error) {
	query := "SELECT id, name, password,active FROM users " +
		"WHERE name = ? AND active = TRUE LIMIT 1"
	return ur.fetch(query, name)
}

func (ur *UserRepository) GetUserByID(id int) (*models.User, error) {
	query := "SELECT id, name, password,active FROM users " +
		"WHERE id = ? LIMIT 1"
	return ur.fetchByID(query, id)
}

func (ur *UserRepository) IsAuthenticated(u *models.User) (bool, error) {
	query := "SELECT id, name, password,active FROM users " +
		"WHERE name = ? AND password = ? AND active = TRUE LIMIT 1"

	fetchedUser, err := ur.fetchByUserNamePassword(query, u)
	if err != nil {
		return false, err
	}
	return *u == *fetchedUser, nil
}

func (ur *UserRepository) fetchByUserNamePassword(query string, user *models.User) (*models.User, error) {
	rows, err := ur.DB.Query(query, user.Name, user.Password)
	if err != nil {
		return nil, err
	}
	//u := &models.User{}
	if rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Password, &user.Active)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, models.ERRNoRecordFound
	}
	return user, nil
}
