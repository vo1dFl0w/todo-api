package store

import "github.com/vo1dFl0w/taskmanager-api/internal/app/store/user"

type Store interface{
	User() user.UserRepository
}