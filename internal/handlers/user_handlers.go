package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/casanera/DlugoshSolutions/internal/models"
	"github.com/casanera/DlugoshSolutions/internal/storage"
)

type UserHandler struct {
	Storage storage.UserStorage
}

func NewUserHandler(s storage.UserStorage) *UserHandler {
	return &UserHandler{Storage: s}
}

// обрабатывает POST-запросы для создания пользователя
// жидает JSON в теле запроса
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Некорректный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if user.Name == "" || user.Email == "" {
		http.Error(w, "Имя и email не могут быть пустыми", http.StatusBadRequest)
		return
	}

	id, err := h.Storage.CreateUser(&user)
	if err != nil {
		log.Printf("Ошибка создания пользователя в хранилище: %v", err)
		http.Error(w, "Внутренняя ошибка сервера при создании пользователя", http.StatusInternalServerError)
		return
	}
	user.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// обрабатывает GET-запросы для получения пользователя по ID
func (h *UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	pathRemainder := strings.TrimPrefix(r.URL.Path, "/api/v1/users")
	idStr := strings.Trim(pathRemainder, "/")

	if idStr != "" { // Если есть ID
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Некорректный ID пользователя '%s': %v", idStr, err)
			http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
			return
		}

		user, err := h.Storage.GetUserByID(id)
		if err != nil {
			if strings.Contains(err.Error(), "пользователь не найден") {
				log.Printf("Пользователь с ID %d не найден: %v", id, err)
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
			} else {
				log.Printf("Ошибка получения пользователя по ID %d из хранилища: %v", id, err)
				http.Error(w, "Внутренняя ошибка сервера при получении пользователя", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	} else {
		users, err := h.Storage.GetAllUsers()
		if err != nil {
			log.Printf("Ошибка получения всех пользователей из хранилища: %v", err)
			http.Error(w, "Внутренняя ошибка сервера при получении списка пользователей", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	pathRemainder := strings.TrimPrefix(r.URL.Path, "/api/v1/users")
	idStr := strings.Trim(pathRemainder, "/")

	if idStr == "" {
		http.Error(w, "ID пользователя должен быть указан в пути для обновления", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Некорректный ID пользователя для обновления '%s': %v", idStr, err)
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	var user models.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Ошибка декодирования JSON при обновлении: %v", err)
		http.Error(w, "Некорректный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	user.ID = id // Устанавливаем ID из URL

	if user.Name == "" || user.Email == "" {
		http.Error(w, "Имя и email не могут быть пустыми при обновлении", http.StatusBadRequest)
		return
	}

	err = h.Storage.UpdateUser(&user)
	if err != nil {
		if strings.Contains(err.Error(), "не найден для обновления") {
			log.Printf("Пользователь с ID %d не найден для обновления: %v", id, err)
			http.Error(w, "Пользователь не найден для обновления", http.StatusNotFound)
		} else {
			log.Printf("Ошибка обновления пользователя с ID %d в хранилище: %v", id, err)
			http.Error(w, "Внутренняя ошибка сервера при обновлении пользователя", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// обрабатывает DELETE-запросы для удаления пользователя
func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	pathRemainder := strings.TrimPrefix(r.URL.Path, "/api/v1/users")
	idStr := strings.Trim(pathRemainder, "/")

	if idStr == "" {
		http.Error(w, "ID пользователя должен быть указан в пути для удаления", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Некорректный ID пользователя для удаления '%s': %v", idStr, err)
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	err = h.Storage.DeleteUser(id)
	if err != nil {
		if strings.Contains(err.Error(), "не найден для удаления") {
			log.Printf("Пользователь с ID %d не найден для удаления: %v", id, err)
			http.Error(w, "Пользователь не найден для удаления", http.StatusNotFound)
		} else {
			log.Printf("Ошибка удаления пользователя с ID %d из хранилища: %v", id, err)
			http.Error(w, "Внутренняя ошибка сервера при удалении пользователя", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
