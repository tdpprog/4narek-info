package main

import (
	"fmt"
	"log"
	"net/http"
)

func textHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем только POST-запросы
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Только POST-запросы")
		return
	}

	// Читаем сырой текст из тела запроса
	text := make([]byte, r.ContentLength)
	r.Body.Read(text)
	
	// Просто возвращаем полученный текст
	w.Header().Set("Content-Type", "text/plain")
	log.Println(w, "Получено: %s", string(text))
}

func main() {
	http.HandleFunc("/buy", textHandler)
	log.Println("Сервер запущен на :8080 (эндпоинт /post-text)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}