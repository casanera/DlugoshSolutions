package models

// структура пользователя в системе
type User struct {
	ID    int64  `json:"id"` // как это поле будет называться при (де)сериализации в JSON
	Name  string `json:"name"`
	Email string `json:"email"`
}
