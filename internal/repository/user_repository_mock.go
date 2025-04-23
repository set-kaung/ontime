package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/set-kaung/senior_project_1/internal/domain"
)

type MockUserRepository struct {
	mutex *sync.RWMutex
	f     *os.File
}

func NewMockUserRepo(f *os.File, mutex *sync.RWMutex) *MockUserRepository {
	return &MockUserRepository{f: f, mutex: mutex}
}

func (m *MockUserRepository) GetUserByID(id int) (*domain.User, error) {
	jsonDecoder := json.NewDecoder(m.f)
	users := []domain.User{}
	m.mutex.RLock()
	err := jsonDecoder.Decode(&users)
	m.mutex.RUnlock()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MockUserRepository) InsertUser(user *domain.User) error {
	jsonDecoder := json.NewDecoder(m.f)
	users := []domain.User{}
	m.mutex.RLock()
	err := jsonDecoder.Decode(&users)
	m.mutex.RUnlock()
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to decode users: %w", err)
	}
	if len(users) > 0 {
		user.ID = users[len(users)-1].ID + 1
	} else {
		user.ID = 1
	}
	users = append(users, *user)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Truncate and seek to beginning
	if err := m.f.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	if _, err := m.f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	jsonEncoder := json.NewEncoder(m.f)
	if err := jsonEncoder.Encode(users); err != nil {
		return fmt.Errorf("failed to encode users: %w", err)
	}

	return nil
}

func (m *MockUserRepository) GetUserByUsername(username string) (*domain.User, error) {
	jsonDecoder := json.NewDecoder(m.f)
	users := []domain.User{}
	m.mutex.RLock()
	err := jsonDecoder.Decode(&users)
	m.mutex.RUnlock()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MockUserRepository) GetUserByEmail(email string) (*domain.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, err := m.f.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek: %w", err)
	}
	jsonDecoder := json.NewDecoder(m.f)
	users := []domain.User{}
	err := jsonDecoder.Decode(&users)
	if err != nil {
		return nil, fmt.Errorf("error getting user by email: %v", err)
	}
	for _, user := range users {
		if user.Email.Address == email {
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}
