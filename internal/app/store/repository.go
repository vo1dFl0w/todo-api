package store

import (
	"time"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
)

type UserRepository interface{
	Create(u *model.User) error
	Find(id int) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	SaveRefreshToken(id int, token string, expiry time.Time) error
	FindByRefreshToken(token string) (*model.User, error)
}