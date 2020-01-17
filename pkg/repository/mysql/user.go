package mysql

import (
	"database/sql"
	slog "github.com/go-eden/slf4go"
	"github.com/ms-clovis/snippetbox/pkg/handlers/validation"
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
	//noinspection ALL
	if rows != nil {
		defer rows.Close()
	}
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

func (ur *UserRepository) fetch(query string, arg string) (*models.User, error) {
	rows, err := ur.DB.Query(query, arg)
	if rows != nil {
		defer rows.Close()
	}
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

func (ur *UserRepository) GetUsers(user *models.User) ([]*models.User, error) {
	query := SELECTSQL +
		" WHERE u.id != ? AND u.active = TRUE"
	rows, err := ur.DB.Query(query, user.ID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]*models.User, 0)

	for rows.Next() {
		u := &models.User{}
		err = rows.Scan(&u.ID, &u.Name, &u.Password, &u.Active)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

var SELECTSQL = "SELECT u.id, u.name, u.password,u.active FROM users u "

//" LEFT JOIN friends f ON u.id = f.watched "

func (ur *UserRepository) GetUser(name string) (*models.User, error) {
	query := SELECTSQL +
		" WHERE u.name = ? AND u.active = TRUE LIMIT 1"
	return ur.fetch(query, name)
}

func (ur *UserRepository) GetUserByID(id int) (*models.User, error) {
	query := SELECTSQL +
		"WHERE u.id = ? AND u.active = TRUE LIMIT 1"
	return ur.fetchByID(query, id)
}

func (ur *UserRepository) IsAuthenticated(hashedPW string, pw string) bool {
	return validation.IsAuthenticated(hashedPW, pw)

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

func (ur *UserRepository) Update(u *models.User) (bool, error) {
	update := "UPDATE snippetbox.users SET name = ?, password = ?, active = ? WHERE id = ?"
	r, err := ur.DB.Exec(update, u.Name, u.Password, u.Active, u.ID)
	if err != nil {
		return false, err
	}
	if rows, err := r.RowsAffected(); rows != 1 || err != nil {
		return false, err
	}
	return true, nil
}
