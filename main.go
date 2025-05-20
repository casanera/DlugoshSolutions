package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var db *sql.DB // глобальная пер. для хранения объекта подключения к БД

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// проверка подлкючения бд
	err := db.Ping()
	if err != nil {
		http.Error(w, "Ошибка подключения к БД: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Сервер работает и успешно подключился к PostgreSQL")
}

func main() {
	dbHost := "localhost"
	dbPort := "5432"
	dbUser := "myuser"
	dbPassword := "mypassword"
	dbName := "mydatabase"

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connStr) // открываем соединение
	if err != nil {
		log.Fatalf("Ошибка при подключении к БД: %v", err)
	}

	err = db.Ping() // пажилая проверка что соединение действительно установлено
	if err != nil {
		log.Fatalf("Не удалось установить соединение с БД: %v", err)
	}
	fmt.Println("Успешное подключение к PostgreSQL")

	http.HandleFunc("/", homeHandler)
	port := ":8080"
	fmt.Printf("Сервер запускается на порту %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
