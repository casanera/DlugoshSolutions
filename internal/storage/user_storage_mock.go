package storage

import (
	"fmt"

	"github.com/casanera/DlugoshSolutions/internal/models"
)

// MockUserStorage является мок-реализацией UserStorage для тестов
type MockUserStorage struct {
	Users         map[int64]*models.User // хранилище пользователей в памяти для мока
	NextID        int64                  // Для генерации ID
	ReturnError   error                  // Какую ошибку возвращать
	UpdateCalled  bool                   // Флаг, что метод UpdateUser был вызван
	DeleteCalled  bool                   // Флаг, что метод DeleteUser был вызван
	GetByIDArg    int64                  // Аргумент, с которым был вызван GetUserByID
	CreateUserArg *models.User           // Аргумент, с которым был вызван CreateUser
}

// создает новый экземпляр MockUserStorage.
func NewMockUserStorage() *MockUserStorage {
	return &MockUserStorage{
		Users:  make(map[int64]*models.User),
		NextID: 1,
	}
}

func (m *MockUserStorage) CreateUser(user *models.User) (int64, error) {
	m.CreateUserArg = user // Сохраняем аргумент для проверки в тесте
	if m.ReturnError != nil {
		return 0, m.ReturnError
	}
	if user.Name == "error_user" {
		return 0, fmt.Errorf("мок: ошибка при создании error_user")
	}
	newID := m.NextID
	m.NextID++
	user.ID = newID
	m.Users[newID] = user
	return newID, nil
}

func (m *MockUserStorage) GetUserByID(id int64) (*models.User, error) {
	m.GetByIDArg = id // Сохраняем аргумент
	if m.ReturnError != nil {
		return nil, m.ReturnError
	}
	user, exists := m.Users[id]
	if !exists {
		return nil, fmt.Errorf("storage.GetUserByID: пользователь не найден")
	}
	return user, nil
}

func (m *MockUserStorage) GetAllUsers() ([]models.User, error) {
	if m.ReturnError != nil {
		return nil, m.ReturnError
	}
	var usersList []models.User
	for _, user := range m.Users {
		usersList = append(usersList, *user)
	}
	return usersList, nil
}

func (m *MockUserStorage) UpdateUser(user *models.User) error {
	m.UpdateCalled = true // Фиксируем вызов
	if m.ReturnError != nil {
		return m.ReturnError
	}
	if _, exists := m.Users[user.ID]; !exists {
		return fmt.Errorf("мок: пользователь с ID %d не найден для обновления", user.ID)
	}
	m.Users[user.ID] = user
	return nil
}

func (m *MockUserStorage) DeleteUser(id int64) error {
	m.DeleteCalled = true // Фиксируем вызов
	if m.ReturnError != nil {
		return m.ReturnError
	}
	if _, exists := m.Users[id]; !exists {
		return fmt.Errorf("мок: пользователь с ID %d не найден для удаления", id)
	}
	delete(m.Users, id)
	return nil
}
