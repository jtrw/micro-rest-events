package repository

import (
	_ "database/sql"
	event "micro-rest-events/internal/repository"
	_ "reflect"

	"github.com/stretchr/testify/mock"
)

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(e event.Event) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockEventRepository) GetOne(uuid string) (event.Event, error) {
	args := m.Called(uuid)
	return args.Get(0).(event.Event), args.Error(1)
}

func (m *MockEventRepository) GetOneByUserId(userId string) (event.Event, error) {
	args := m.Called(userId)
	return args.Get(0).(event.Event), args.Error(1)
}

func (m *MockEventRepository) GetAll(q event.Query) ([]event.Event, error) {
	args := m.Called(q)
	return args.Get(0).([]event.Event), args.Error(1)
}

func (m *MockEventRepository) GetAllByUserId(userId string, q event.Query) ([]event.Event, error) {
	args := m.Called(userId, q)
	return args.Get(0).([]event.Event), args.Error(1)
}

func (m *MockEventRepository) Count(q event.Query) (int, error) {
	args := m.Called(q)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockEventRepository) CountByUserId(userId string, q event.Query) (int, error) {
	args := m.Called(userId, q)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockEventRepository) ChangeStatus(uuid string, e event.Event) (int64, error) {
	args := m.Called(uuid, e)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepository) ChangeIsSeen(uuid string) (int64, error) {
	args := m.Called(uuid)
	return args.Get(0).(int64), args.Error(1)
}
