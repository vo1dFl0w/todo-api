package model

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

type Task struct {
	UserID      int        `json:"user_id"`
	TaskID      int        `json:"task_id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Deadline    *time.Time `json:"deadline"`
	Complete    *bool      `json:"complete"`
}

type CustomTime struct {
	time.Time
}

const timeLayout = "2006-01-02 15:04:05"

var (
	getTask = http.MethodGet
	createTask = http.MethodPost
	updateTask = http.MethodPatch
	deleteTask = http.MethodDelete
)

func (t *Task) Validation(method string) error {

	switch method {
	case createTask:
		if t.Title == nil || *t.Title == "" {
			return errors.New("title cannot be empty")
		}
	
		if t.Deadline == nil || time.Now().After(*t.Deadline) {
			return errors.New("wrong format of deadline: use format YYYY-MM-DD HH:MM:SS")
		}
	case updateTask:
		if t.Title != nil && *t.Title == "" {
			return errors.New("title cannot be empty")
		}
	
		if t.Deadline != nil && time.Now().After(*t.Deadline) {
			return errors.New("deadline must be in the future")
		}
	}

	return nil
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	parsedTime, err := time.Parse(timeLayout, s)
	if err != nil {
		return errors.New("invalid time fromat: use YYYY-MM-DD HH:MM:SS")
	}

	ct.Time = parsedTime

	return nil
}
