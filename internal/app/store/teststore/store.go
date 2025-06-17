package teststore

import (
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/todo"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/user"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/teststore/todo_teststore"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/teststore/user_teststore"
)

type Store struct {
	userRepository user.UserRepository
	todoRepository todo.TodoRepository
}

func New() *Store {
	return &Store{}
}

func (s *Store) User() user.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}
	
	s.userRepository = &user_teststore.UserRepository{
		Users: make(map[int]*model.User),
	}

	return s.userRepository
}

func (s *Store) Todo() todo.TodoRepository {
	if s.todoRepository != nil {
		return s.todoRepository
	}
	
	s.todoRepository = &todo_teststore.TodoRepository{
		Tasks: make(map[int]map[string]string),
	}

	return s.todoRepository
}