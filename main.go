package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "Подключение к БД не инициализировано", http.StatusInternalServerError)
		return
	}
	err := db.Ping()
	if err != nil {
		http.Error(w, "Ошибка подключения к БД: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Ошибка Ping к БД: %v", err) // Логируем ошибку
		return
	}
	fmt.Fprintf(w, "Cервер работает и успешно подключился к PostgreSQL")
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Проверка, что все необходимые переменные окружения установлены
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Одна или несколько переменных окружения для БД не установлены (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка при подключении к БД: %v", err)
	}

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			break // Соединение успешно
		}
		log.Printf("Не удалось подключиться к БД (попытка %d/%d): %v. Ожидание...", i+1, maxRetries, err)
	}
	if err != nil {
		log.Fatalf("Не удалось установить соединение с БД после %d попыток: %v", maxRetries, err)
	}

	fmt.Println("Успешное подключение к PostgreSQL из контейнера backend")

	http.HandleFunc("/", homeHandler)
	appPort := os.Getenv("APP_PORT") // Если порт приложения тоже из переменной
	if appPort == "" {
		appPort = "8080" // Значение по умолчанию
	}

	fmt.Printf("Сервер backend запускается на порту :%s\n", appPort)
	if err := http.ListenAndServe(":"+appPort, nil); err != nil {
		log.Fatal(err)
	}
}
