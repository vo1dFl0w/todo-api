package user_teststore

import (
	"time"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

type UserRepository struct {
	Store *store.Store
	Users map[int]*model.User
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validation(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	u.ID = len(r.Users) + 1
	r.Users[u.ID] = u
		
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	for _, u := range r.Users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	u, ok := r.Users[id]
	if !ok {
		return nil, store.ErrRecordNotFound
	}

	return u, nil
}

func (r *UserRepository) GetRefreshTokenExpire(token string) (*model.User, error) {
	for _, u := range r.Users {
		if u.RefreshToken == token {
			return u, nil
		}
	}
	
	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) SaveRefreshToken(id int, token string, expiry time.Time) error {
	u, ok := r.Users[id]
	if !ok {
		return store.ErrRecordNotFound
	}

	u.RefreshToken = token
	u.RefreshTokenExpire = expiry

	return nil
}