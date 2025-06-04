package dbstore

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

type Store struct {
	db 				*sql.DB
	userReposiotry 	*UserReposiotry
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) User() store.UserRepository {
	if s.userReposiotry != nil {
		return s.userReposiotry
	}

	s.userReposiotry = &UserReposiotry{
		store: s,
	}

	return s.userReposiotry
}


