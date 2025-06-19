package todo_postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
)

type TodoRepository struct {
	DB *sql.DB
}

func (r *TodoRepository) Get(id int) (*model.Task, error) {
	t := &model.Task{}
	
	rows, err := r.DB.Query(
		"SELECT user_id, task_id, title, description, deadline, complete FROM tasks WHERE user_id = $1",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&t.UserID,
			&t.TaskID,
			&t.Title,
			&t.Description,
			&t.Deadline,
			&t.Complete,
		)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, sql.ErrNoRows
	}

	return t, nil
}

func (r *TodoRepository) Create(t *model.Task) error {
	if err := r.DB.QueryRow(
		"INSERT INTO tasks (user_id, title, description, deadline, complete) VALUES ($1, $2, $3, $4, $5) RETURNING task_id",
		t.UserID,
		t.Title,
		t.Description,
		t.Deadline,
		t.Complete,
	).Scan(&t.TaskID); err != nil {
		return err
	}

	return nil
}

func (r *TodoRepository) Update(t *model.Task) error {
	query := "UPDATE tasks SET "
	args := []interface{}{}
	placeholders := []string{}
	i := 1

	if t.Title != nil {
		placeholders = append(placeholders, fmt.Sprintf("title = $%d", i))
		args = append(args, *t.Title)
		i++
	}

	if t.Description != nil {
		placeholders = append(placeholders, fmt.Sprintf("description = $%d", i))
		args = append(args, *t.Description)
		i++
	}

	if t.Deadline != nil {
		placeholders = append(placeholders, fmt.Sprintf("deadline = $%d", i))
		args = append(args, *t.Deadline)
		i++
	}

	if t.Complete != nil {
		placeholders = append(placeholders, fmt.Sprintf("complete = $%d", i))
		args = append(args, *t.Complete)
		i++
	}

	if len(placeholders) == 0 {
		return nil
	}

	query += strings.Join(placeholders, ", ")
	query += fmt.Sprintf(" WHERE task_id = $%d AND user_id = $%d", i, i+1)
	args = append(args, t.TaskID, t.UserID)

	_, err := r.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *TodoRepository) Delete(userID int, taskIDs []int) (int64, error) {
	if len(taskIDs) == 0 {
		return 0, nil
	}

	query := `DELETE FROM tasks WHERE user_id = $1 AND task_id = ANY($2)`
	
	res, err := r.DB.Exec(query, userID, pq.Array(taskIDs))
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *TodoRepository) FindTaskByTaskID(userID int, taskIDs []int) error {
	rows, err := r.DB.Query(
		"SELECT task_id FROM tasks WHERE user_id = $1 AND task_id = ANY($2)",
		userID,
		pq.Array(taskIDs),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	foundIDs := map[int]bool{}

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		foundIDs[id] = true
	}

	for _, id := range taskIDs {
		if !foundIDs[id] {
			return fmt.Errorf("task with id %d not found", id)
		}
	}

	return nil
}
