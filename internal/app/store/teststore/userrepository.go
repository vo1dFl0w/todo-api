package teststore

import (
	"time"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

type UserRepository struct {
	store *Store
	users map[int]*model.User
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validation(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	r.users[u.ID] = u
	u.ID = len(r.users)
		
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, store.ErrRecordNotFound
	}

	return u, nil
}

func (r *UserRepository) FindByRefreshToken(token string) (*model.User, error) {
	for _, u := range r.users {
		if u.RefreshToken == token {
			return u, nil
		}
	}
	
	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) SaveRefreshToken(id int, token string, expiry time.Time) error {
	r.users[id] = &model.User{
		RefreshToken: token,
		RefreshTokenExpire: expiry,
	}
	
	return nil
}