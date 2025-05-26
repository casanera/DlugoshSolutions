package handlers

import (
	"bytes"         // Для создания io.Reader из строки (тело запроса)
	"encoding/json" // Для кодирования/декодирования JSON
	"net/http"
	"net/http/httptest" // Для тестирования HTTP обработчиков
	"strconv"
	"testing" // Пакет для написания тестов

	"fmt"

	"github.com/casanera/DlugoshSolutions/internal/models"
	"github.com/casanera/DlugoshSolutions/internal/storage"
)

func TestCreateUserHandler(t *testing.T) {
	mockStorage := storage.NewMockUserStorage()
	userHandler := NewUserHandler(mockStorage) // Создаем хендлер с моком

	// Подготовка тестовых случаев
	testCases := []struct {
		name               string
		inputBody          string
		expectedStatusCode int // Ожидаемый HTTP статус-код
		expectedName       string
		expectErrorInBody  bool
		setupMock          func() // Функция для настройки мока перед тестом
	}{
		{
			name:               "Успешное создание",
			inputBody:          `{"name": "Test User", "email": "test@example.com"}`,
			expectedStatusCode: http.StatusCreated,
			expectedName:       "Test User",
		},
		{
			name:               "Некорректный JSON",
			inputBody:          `{"name": "Test User", "email": "test@example.com"`, // нет закрывающей скобки
			expectedStatusCode: http.StatusBadRequest,
			expectErrorInBody:  true,
		},
		{
			name:               "Пустое имя",
			inputBody:          `{"name": "", "email": "empty@example.com"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectErrorInBody:  true,
		},
		{
			name:      "Ошибка хранилища при создании",
			inputBody: `{"name": "DB Error User", "email": "dberror@example.com"}`,
			setupMock: func() {
				mockStorage.ReturnError = fmt.Errorf("симуляция ошибки БД")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectErrorInBody:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Сбрасываем состояние мока перед каждым тестом
			mockStorage.ReturnError = nil
			mockStorage.Users = make(map[int64]*models.User) // Очищаем пользователей
			mockStorage.NextID = 1

			if tc.setupMock != nil {
				tc.setupMock() // Настраиваем мок, если нужно для этого теста
			}

			// Создаем HTTP запрос
			req, err := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(tc.inputBody))
			if err != nil {
				t.Fatalf("Не удалось создать запрос: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			userHandler.CreateUserHandler(rr, req)

			// Проверяем статус-код
			if status := rr.Code; status != tc.expectedStatusCode {
				t.Errorf("Обработчик вернул неверный статус-код: получено %v, ожидалось %v. Тело ответа: %s",
					status, tc.expectedStatusCode, rr.Body.String())
			}

			if tc.expectedStatusCode == http.StatusCreated && !tc.expectErrorInBody {
				var returnedUser models.User
				if err := json.NewDecoder(rr.Body).Decode(&returnedUser); err != nil {
					t.Fatalf("Не удалось декодировать JSON ответа: %v. Тело: %s", err, rr.Body.String())
				}
				if returnedUser.Name != tc.expectedName {
					t.Errorf("Имя пользователя в ответе неверное: получено '%s', ожидалось '%s'",
						returnedUser.Name, tc.expectedName)
				}
				if returnedUser.ID == 0 {
					t.Errorf("ID пользователя в ответе не должен быть 0")
				}
			}

			if tc.expectErrorInBody && rr.Body.String() == "" && (tc.expectedStatusCode >= 400) {
				t.Errorf("Ожидалось тело с ошибкой, но получено пустое тело ответа")
			}
		})
	}
}

func TestGetUserHandler(t *testing.T) {
	mockStorage := storage.NewMockUserStorage()
	userHandler := NewUserHandler(mockStorage)
	existingUser := &models.User{ID: 1, Name: "Existing User", Email: "existing@example.com"}
	mockStorage.Users[existingUser.ID] = existingUser
	mockStorage.NextID = 2

	t.Run("Получение всех пользователей (один существует)", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/", nil)
		rr := httptest.NewRecorder()
		userHandler.GetUserHandler(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GetAll: неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusOK, rr.Body.String())
		}

		var users []models.User
		if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
			t.Fatalf("GetAll: не удалось декодировать JSON: %v. Тело: %s", err, rr.Body.String())
		}
		if len(users) != 1 {
			t.Errorf("GetAll: ожидался 1 пользователь, получено %d", len(users))
		}
		if len(users) > 0 && users[0].Name != existingUser.Name {
			t.Errorf("GetAll: имя пользователя неверное: получено '%s', ожидалось '%s'", users[0].Name, existingUser.Name)
		}
	})

	t.Run("Получение пользователя по существующему ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/"+strconv.FormatInt(existingUser.ID, 10), nil)
		rr := httptest.NewRecorder()
		userHandler.GetUserHandler(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GetByID: неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusOK, rr.Body.String())
		}
		var user models.User
		if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
			t.Fatalf("GetByID: не удалось декодировать JSON: %v. Тело: %s", err, rr.Body.String())
		}
		if user.Name != existingUser.Name {
			t.Errorf("GetByID: имя пользователя неверное: получено '%s', ожидалось '%s'", user.Name, existingUser.Name)
		}
	})

	t.Run("Получение пользователя по несуществующему ID", func(t *testing.T) {
		nonExistentID := int64(999)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/"+strconv.FormatInt(nonExistentID, 10), nil)
		rr := httptest.NewRecorder()
		userHandler.GetUserHandler(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("GetByID (not found): неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusNotFound, rr.Body.String())
		}
	})

	t.Run("Получение пользователя с некорректным ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)
		rr := httptest.NewRecorder()
		userHandler.GetUserHandler(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("GetByID (invalid id): неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusBadRequest, rr.Body.String())
		}
	})
}
