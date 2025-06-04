package store

import "github.com/vo1dFl0w/taskmanager-api/internal/app/model"

type UserRepository interface{
	Create(u *model.User) error
	Find(id int) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
}