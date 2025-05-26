package storage

import (
	"database/sql"
	"fmt"

	"github.com/casanera/DlugoshSolutions/internal/models"
)

type UserStorage interface {
	CreateUser(user *models.User) (int64, error)
	GetUserByID(id int64) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int64) error
}

type PostgresUserStorage struct {
	DB *sql.DB
}

func NewPostgresUserStorage(db *sql.DB) *PostgresUserStorage {
	return &PostgresUserStorage{DB: db}
}

// CreateUser добавляет нового пользователя в базу данных
// Возвращает ID созданного пользователя или ошибку
func (s *PostgresUserStorage) CreateUser(user *models.User) (int64, error) {
	query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
	var id int64
	err := s.DB.QueryRow(query, user.Name, user.Email).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("storage.CreateUser: %w", err)
	}
	return id, nil
}

// GetUserByID получает пользователя по ID
func (s *PostgresUserStorage) GetUserByID(id int64) (*models.User, error) {
	query := "SELECT id, name, email FROM users WHERE id = $1"
	user := &models.User{}
	err := s.DB.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows { // спец ошибка, если запись не найдена
			return nil, fmt.Errorf("storage.GetUserByID: пользователь не найден: %w", err)
		}
		return nil, fmt.Errorf("storage.GetUserByID: %w", err)
	}
	return user, nil
}

// получает всех пользователей.
func (s *PostgresUserStorage) GetAllUsers() ([]models.User, error) {
	query := "SELECT id, name, email FROM users ORDER BY id ASC"
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("storage.GetAllUsers: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, fmt.Errorf("storage.GetAllUsers: ошибка сканирования строки: %w", err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("storage.GetAllUsers: ошибка после итерации: %w", err)
	}

	return users, nil
}

func (s *PostgresUserStorage) UpdateUser(user *models.User) error {
	query := "UPDATE users SET name = $1, email = $2 WHERE id = $3"
	result, err := s.DB.Exec(query, user.Name, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("storage.UpdateUser: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage.UpdateUser: не удалось получить количество измененных строк: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("storage.UpdateUser: пользователь с ID %d не найден для обновления", user.ID)
	}
	return nil
}

func (s *PostgresUserStorage) DeleteUser(id int64) error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := s.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("storage.DeleteUser: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage.DeleteUser: не удалось получить количество удаленных строк: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("storage.DeleteUser: пользователь с ID %d не найден для удаления", id)
	}
	return nil
}
