package main

import (
	"fmt"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Тестирование сервера на Go")
}

func main() {
	http.HandleFunc("/", homeHandler)

	// порт, на котором будет работать сервер
	port := ":8080"
	fmt.Printf("Сервер запускается на порту %s\n", port)

	// сообщение об ошибке, если возникнет
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
