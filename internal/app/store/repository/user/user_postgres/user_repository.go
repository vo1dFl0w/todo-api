package user_postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

type UserReposiotry struct {
	DB *sql.DB
}

func (r *UserReposiotry) Create(u *model.User) error {
	if err := u.Validation(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	err := r.DB.QueryRow(
		"INSERT INTO users (email, encrypted_password) VALUES ($1, $2) RETURNING id",
		u.Email,
		u.EncryptedPassword,
	).Scan(&u.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			err = errors.New("a user with this email already exists")
			return err
		}
	}
	
	return nil
}

func (r *UserReposiotry) FindByID(id int) (*model.User, error) {
	u := &model.User{}

	if err := r.DB.QueryRow(
		"SELECT id, email, encrypted_password, refresh_token, refresh_token_exp FROM users WHERE id = $1",
		id,
	).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
		&u.RefreshToken,
		&u.RefreshTokenExpire,
	); err != nil {
		return nil, store.ErrRecordNotFound
	}
	
	return u, nil
}

func (r *UserReposiotry) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}

	if err := r.DB.QueryRow(
		"SELECT id, email, encrypted_password FROM users WHERE email = $1",
		email,
	).Scan(&u.ID, &u.Email, &u.EncryptedPassword); err != nil{
		return nil, store.ErrRecordNotFound
	}

	return u, nil
}

func (r *UserReposiotry) SaveRefreshToken(id int, token string, expiry time.Time) error {
	_, err := r.DB.Exec(
		"UPDATE users SET refresh_token = $1, refresh_token_exp = $2 WHERE id = $3",
		token,
		expiry,
		id,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *UserReposiotry) GetRefreshTokenExpire(token string) (*model.User, error) {
	u := &model.User{}

	if err := r.DB.QueryRow(
		"SELECT id, refresh_token_exp FROM users WHERE refresh_token = $1",
		token,
	).Scan(&u.ID, &u.RefreshTokenExpire); err != nil {
		return nil, store.ErrRecordNotFound
	}

	return u, nil
}