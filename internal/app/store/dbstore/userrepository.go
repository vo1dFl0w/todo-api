package dbstore

import (
	"errors"

	"github.com/lib/pq"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
)

type UserReposiotry struct {
	store *Store
}

func (r *UserReposiotry) Create(u *model.User) error {
	if err := u.Validation(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	err := r.store.db.QueryRow(
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

func (r *UserReposiotry) Find(id int) (*model.User, error) {
	return nil, nil
}

func (r *UserReposiotry) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}

	if err := r.store.db.QueryRow(
		"SELECT id, email, encrypted_password FROM users WHERE email = $1",
		email,
	).Scan(&u.ID, &u.Email, &u.EncryptedPassword); err != nil{
		return nil, err
	}

	return u, nil
}