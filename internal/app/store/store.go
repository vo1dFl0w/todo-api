package store

import (
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/todo"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/user"
)

type Store interface{
	User() user.UserRepository
	Todo() todo.TodoRepository
}