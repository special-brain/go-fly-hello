package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	// Мьютекс для безопасного доступа к массиву из разных горутин
	mu           sync.Mutex
	lastRequests []time.Time
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Блокируем доступ к массиву для других запросов
	mu.Lock()
	
	// Добавляем текущее время в начало массива
	lastRequests = append([]time.Time{time.Now()}, lastRequests...)
	
	// Оставляем только последние 5 записей
	if len(lastRequests) > 5 {
		lastRequests = lastRequests[:5]
	}
	
	// Делаем копию массива для безопасного вывода
	requestsCopy := make([]time.Time, len(lastRequests))
	copy(requestsCopy, lastRequests)
	
	// Снимаем блокировку, чтобы другие запросы не ждали
	mu.Unlock()

	// Отправляем ответ клиенту
	fmt.Fprintln(w, "Hello, World!")
	fmt.Fprintln(w, "\nВремя последних 5 обращений:")
	for i, t := range requestsCopy {
		fmt.Fprintf(w, "%d. %s\n", i+1, t.Format(time.RFC1123))
	}
}

func main() {
	http.HandleFunc("/", helloHandler)
	
	fmt.Println("Сервер запущен на порту 8080...")
	// Fly.io ожидает, что приложение будет слушать порт 8080 (по умолчанию)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Ошибка запуска сервера: %s\n", err)
	}
}