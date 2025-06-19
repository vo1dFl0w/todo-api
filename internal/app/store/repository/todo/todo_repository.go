package todo

import "github.com/vo1dFl0w/taskmanager-api/internal/app/model"

type TodoRepository interface{
	Get(int) (*model.Task, error)
	Create(*model.Task) error
	Update(*model.Task) error
	Delete(int, []int) (int64, error)
}