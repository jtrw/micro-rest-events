package repository

import (
    "database/sql"
	"github.com/stretchr/testify/mock"
	event "micro-rest-events/v1/app/backend/repository"
)

type MockEventRepository_OLD struct {
	mock.Mock
}

func NewEventRepository_OLD(conn *sql.DB) *MockEventRepository_OLD {
	return &MockEventRepository_OLD{}
}

func (m *MockEventRepository_OLD) Create(e event.Event) error {
	return nil
}

func (m *MockEventRepository_OLD) GetOne(uuid string) (event.Event, error) {
	args := m.Called(uuid)
	return args.Get(0).(event.Event), args.Error(1)
}

func (m *MockEventRepository_OLD) GetByUserId(userId int) (event.Event, error) {
	args := m.Called(userId)
	return args.Get(0).(event.Event), args.Error(1)
}

func (m *MockEventRepository_OLD) ChangeStatus(uuid string, e event.Event) (int64, error) {
	args := m.Called(uuid, e)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepository_OLD) ChangeIsSeen(uuid string) (int64, error) {
	args := m.Called(uuid)
	return args.Get(0).(int64), args.Error(1)
}
