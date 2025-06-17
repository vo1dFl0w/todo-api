package todo_teststore

type TodoRepository struct {
	Tasks map[int]map[string]string
}

func (r *TodoRepository) CreateTask() error {
	return nil
}