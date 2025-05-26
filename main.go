package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq" // Драйвер PostgreSQL

	"github.com/casanera/DlugoshSolutions/internal/handlers"
	"github.com/casanera/DlugoshSolutions/internal/storage"
)

var db *sql.DB // Глобальная переменная для хранения объекта подключения к БД

func routeHandler(userH *handlers.UserHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Входящий запрос (через routeHandler): Метод=%s, Путь=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr) // Добавлен идентификатор

		// Обработка API эндпоинтов для пользователей
		if strings.HasPrefix(r.URL.Path, "/api/v1/users") {
			pathRemainder := strings.TrimPrefix(r.URL.Path, "/api/v1/users")
			isSpecificUser := pathRemainder != "" && pathRemainder != "/"

			switch r.Method {
			case http.MethodGet:
				userH.GetUserHandler(w, r)
			case http.MethodPost:
				if !isSpecificUser || pathRemainder == "/" {
					userH.CreateUserHandler(w, r)
				} else {
					log.Printf("Некорректный POST запрос на %s", r.URL.Path)
					http.Error(w, "Метод POST применим только к /api/v1/users", http.StatusMethodNotAllowed)
				}
			case http.MethodPut:
				// PUT запросы для обновления пользователя требуют ID
				if isSpecificUser {
					userH.UpdateUserHandler(w, r)
				} else {
					log.Printf("PUT запрос без ID пользователя на %s", r.URL.Path)
					http.Error(w, "Для PUT запроса требуется ID пользователя в пути (например, /api/v1/users/123)", http.StatusBadRequest)
				}
			case http.MethodDelete:
				// DELETE запросы для удаления пользователя требуют ID
				if isSpecificUser {
					userH.DeleteUserHandler(w, r)
				} else {
					log.Printf("DELETE запрос без ID пользователя на %s", r.URL.Path)
					http.Error(w, "Для DELETE запроса требуется ID пользователя в пути (например, /api/v1/users/123)", http.StatusBadRequest)
				}
			default:
				log.Printf("Метод %s не разрешен для %s", r.Method, r.URL.Path)
				http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			}
			return
		}
		log.Printf("routeHandler: Маршрут не найден для: %s (это неожиданно, если не /api/)", r.URL.Path)
		http.NotFound(w, r)
	}
}

// homeHandler больше не будет напрямую обрабатывать "/", так как это делает FileServer.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		log.Println("Ошибка в homeHandler: подключение к БД (db) равно nil")
		http.Error(w, "Внутренняя ошибка сервера: подключение к БД не инициализировано.", http.StatusInternalServerError)
		return
	}
	err := db.Ping()
	if err != nil {
		log.Printf("Ошибка Ping к БД в homeHandler: %v", err)
		http.Error(w, "Внутренняя ошибка сервера: не удалось подключиться к БД. "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Статус сервера: ОК. Подключение к PostgreSQL: ОК. Фронтенд доступен по корневому пути '/'. API на '/api/v1/users'.")
}

func main() {
	log.Println("Запуск приложения...")

	// Получаем конфигурацию для подключения к БД из переменных окружения
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		missingVars := []string{}
		if dbHost == "" {
			missingVars = append(missingVars, "DB_HOST")
		}
		if dbPort == "" {
			missingVars = append(missingVars, "DB_PORT")
		}
		if dbUser == "" {
			missingVars = append(missingVars, "DB_USER")
		}
		if dbPassword == "" {
			missingVars = append(missingVars, "DB_PASSWORD")
		}
		if dbName == "" {
			missingVars = append(missingVars, "DB_NAME")
		}
		log.Fatalf("Критической(-ие) переменной(-ые) окружения для БД не установлена(-ы): %s. Завершение работы.", strings.Join(missingVars, ", "))
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Println("Попытка подключения к PostgreSQL...")
	var err error
	// Открываем соединение с базой данных и присваиваем его глобальной переменной db
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка при вызове sql.Open для PostgreSQL: %v", err)
	}

	// пытаемся подключиться к БД несколько раз с задержкой,
	// так как контейнер с БД может стартовать медленнее, чем приложение
	maxRetries := 15
	for i := 0; i < maxRetries; i++ {
		log.Printf("Проверка соединения с БД (попытка %d/%d)...", i+1, maxRetries)
		err = db.Ping()
		if err == nil {
			log.Println("Успешное подключение к PostgreSQL!")
			break
		}
		log.Printf("Не удалось подключиться к БД: %v. Ожидание 5 секунд перед следующей попыткой...", err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Не удалось установить соединение с PostgreSQL после %d попыток: %v. Завершение работы.", maxRetries, err)
	}

	userStorage := storage.NewPostgresUserStorage(db)
	userHandler := handlers.NewUserHandler(userStorage)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", routeHandler(userHandler))
	staticFileServer := http.FileServer(http.Dir("./static"))
	// Регистрируем FileServer для обработки всех запросов к корневому пути "/" и его подпутям
	mux.Handle("/", staticFileServer)
	mux.HandleFunc("/status", homeHandler)
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080" // Порт по умолчанию
		log.Printf("Переменная окружения APP_PORT не установлена, используется порт по умолчанию: %s", appPort)
	}
	log.Printf("Сервер backend (с CRUD и фронтендом) запускается на порту :%s", appPort)
	log.Printf("API пользователей доступно по /api/v1/users (обрабатывается через routeHandler)")
	log.Printf("Статические файлы (фронтенд) раздаются из папки ./static и доступны по адресу: http://localhost:%s/", appPort)
	log.Printf("Для проверки статуса (если раскомментирован /status): http://localhost:%s/status", appPort)

	// Вместо `http.ListenAndServe(":"+appPort, nil)` используем `mux`.
	if err := http.ListenAndServe(":"+appPort, mux); err != nil {
		log.Fatalf("Ошибка при запуске HTTP-сервера: %v", err)
	}
}
