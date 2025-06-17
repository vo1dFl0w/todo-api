package repository

import (
	"database/sql"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/todo"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/todo/todo_postgres"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/user"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository/user/user_postgres"
)

type Store struct {
	DB *sql.DB
	userRepository user.UserRepository
	todoRepository todo.TodoRepository
}

func New(db *sql.DB) *Store{
	return &Store{
		DB: db,
	}
}

func (s *Store) User() user.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &user_postgres.UserReposiotry{
		DB: s.DB,
	}

	return s.userRepository
}

func (s *Store) Todo() todo.TodoRepository {
	if s.todoRepository != nil {
		return s.todoRepository
	}

	s.todoRepository = &todo_postgres.TodoRepository{
		DB: s.DB,
	}

	return s.todoRepository
}