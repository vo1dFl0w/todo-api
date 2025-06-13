package teststore_user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
	teststore_user "github.com/vo1dFl0w/taskmanager-api/internal/app/store/teststore/user"
)

func TestUserRepository_Create(t *testing.T) {
	s := teststore_user.New()

	u1 := &model.User{}
	err := s.User().Create(u1)

	assert.Error(t, err)

	u2 := model.TestUser(t)
	err = s.User().Create(u2)

	assert.NoError(t, err)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	s := teststore_user.New()
	email := "user@example.org"

	_, err := s.User().FindByEmail(email)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	u := model.TestUser(t)
	u.Email = email
	s.User().Create(u)

	u, err = s.User().FindByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}
