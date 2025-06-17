package todo

type TodoRepository interface{
	CreateTask() error
}