package todo_postgres

import (
	"database/sql"
)

type TodoRepository struct {
	DB *sql.DB
}

func (r *TodoRepository) CreateTask() error {
	return nil
}
