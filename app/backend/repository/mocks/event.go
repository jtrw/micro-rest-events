package repository

import (
    "database/sql"
	"github.com/stretchr/testify/mock"
	event "micro-rest-events/v1/app/backend/repository"
)

type MockEventRepository struct {
	mock.Mock
}

func NewEventRepository(conn *sql.DB) *MockEventRepository {
	return &MockEventRepository{}
}

func (m *MockEventRepository) Create(e event.Event) error {
	return nil
}

func (m *MockEventRepository) GetOne(uuid string) (event.Event, error) {
	args := m.Called(uuid)
	return args.Get(0).(event.Event), args.Error(1)
}

func (m *MockEventRepository) GetByUserId(userId int) (event.Event, error) {
	args := m.Called(userId)
	return args.Get(0).(event.Event), args.Error(1)
}

func (m *MockEventRepository) ChangeStatus(uuid string, e event.Event) (int64, error) {
	args := m.Called(uuid, e)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepository) ChangeIsSeen(uuid string) (int64, error) {
	args := m.Called(uuid)
	return args.Get(0).(int64), args.Error(1)
}
