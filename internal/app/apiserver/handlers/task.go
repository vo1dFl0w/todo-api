package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/middleware"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/services"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

type TaskHandler struct {
	Store        store.Store
	TokenService services.TokenService
	Respond      func(http.ResponseWriter, *http.Request, int, interface{})
	Error        func(http.ResponseWriter, *http.Request, int, error)
}

func (h *TaskHandler) GetTask(userID int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			h.Error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		authUser := r.Context().Value(middleware.CtxKeyUser).(*model.User)

		if authUser.ID != userID {
			h.Error(w, r, http.StatusForbidden, errors.New("access denied"))
			return
		}

		tasks, err := h.Store.Todo().Get(userID)
		if err != nil {
			h.Error(w, r, http.StatusNotFound, err)
			return
		}

		h.Respond(w, r, http.StatusOK, tasks)
	}
}

func (h *TaskHandler) CreateTask(userID int) http.HandlerFunc {
	type request struct {
		UserID      int              `json:"user_id"`
		Title       string           `json:"title"`
		Description string           `json:"description"`
		Deadline    model.CustomTime `json:"deadline"`
		Complete    bool            `json:"complete,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			h.Error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		authUser := r.Context().Value(middleware.CtxKeyUser).(*model.User)

		if authUser.ID != userID {
			h.Error(w, r, http.StatusForbidden, errors.New("access denied"))
			return
		}

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			h.Error(w, r, http.StatusBadRequest, err)
			return
		}

		t := &model.Task{
			UserID:      userID,
			Title:       &req.Title,
			Description: &req.Description,
			Deadline:    &req.Deadline.Time,
			Complete:    &req.Complete,
		}

		if err := t.Validation(r.Method); err != nil {
			h.Error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		if err := h.Store.Todo().Create(t); err != nil {
			h.Error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		h.Respond(w, r, http.StatusCreated, t)
	}
}

func (h *TaskHandler) UpdateTask(userID int, taskID int) http.HandlerFunc {
	type request struct {
		Title       *string           `json:"title,omitempty"`
		Description *string           `json:"description,omitempty"`
		Deadline    *model.CustomTime `json:"deadline,omitempty"`
		Complete    *bool            `json:"complete,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			h.Error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		authUser := r.Context().Value(middleware.CtxKeyUser).(*model.User)

		if authUser.ID != userID {
			h.Error(w, r, http.StatusForbidden, errors.New("access denied"))
			return
		}

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			h.Error(w, r, http.StatusBadRequest, err)
			return
		}

		t := &model.Task{
			UserID: userID,
			TaskID: taskID,
		}

		if req.Title != nil {
			t.Title = req.Title
		}

		if req.Description != nil {
			t.Description = req.Description
		}

		if req.Deadline != nil {
			t.Deadline = &req.Deadline.Time
		}

		if req.Complete != nil {
			t.Complete = req.Complete
		}

		if err := t.Validation(r.Method); err != nil {
			h.Error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		
		if err := h.Store.Todo().Update(t); err != nil {
			h.Error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		h.Respond(w, r, http.StatusOK, nil)
	}
}

func (h *TaskHandler) DeleteTask(userID int, taskIDs []int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			h.Error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		authUser := r.Context().Value(middleware.CtxKeyUser).(*model.User)

		if authUser.ID != userID {
			h.Error(w, r, http.StatusForbidden, errors.New("access denied"))
			return
		}

		

		count, err := h.Store.Todo().Delete(userID, taskIDs)
		if err != nil {
			h.Error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		h.Respond(w, r, http.StatusOK, count)
	}
}
