package todo_teststore

import "github.com/vo1dFl0w/taskmanager-api/internal/app/model"

type TodoRepository struct {
	Tasks map[int]map[string]string
}

func (r *TodoRepository) Get(int) (*model.Task, error) {
	t := &model.Task{}

	return t, nil
}

func (r *TodoRepository) Create(*model.Task) error {
	return nil
}

func (r *TodoRepository) Update(t *model.Task) error {
	return nil
}

func (r *TodoRepository) Delete(userID int, taskID []int) (int64, error) {
	return 0, nil
}