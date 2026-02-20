package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite" // Подключаем драйвер SQLite на чистом Go
)

var db *sql.DB

func initDB() {
	var err error
	
	// Если папки /data нет (например, при локальном тесте), создадим папку data рядом с файлом
	dbPath := "/data/requests.db"
	if _, err := os.Stat("/data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
		dbPath = "data/requests.db"
	}

	// Открываем или создаем файл БД
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Ошибка открытия БД:", err)
	}

	// Создаем таблицу, если её еще нет
	query := `
	CREATE TABLE IF NOT EXISTS requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Ошибка создания таблицы:", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Записываем текущий запрос в БД (в UTC)
	_, err := db.Exec("INSERT INTO requests (timestamp) VALUES (?)", time.Now().UTC())
	if err != nil {
		http.Error(w, "Ошибка записи в БД", 500)
		return
	}

	// 2. Читаем последние 5 записей (сортировка по убыванию ID)
	rows, err := db.Query("SELECT timestamp FROM requests ORDER BY id DESC LIMIT 5")
	if err != nil {
		http.Error(w, "Ошибка чтения из БД", 500)
		return
	}
	defer rows.Close()

	// 3. Выводим результат
	fmt.Fprintln(w, "Hello, Persistent World!")
	fmt.Fprintln(w, "\nВремя последних 5 обращений (из SQLite):")
	
	i := 1
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			continue
		}
		// Форматируем время для красоты
		fmt.Fprintf(w, "%d. %s\n", i, t.Format(time.RFC1123))
		i++
	}
}

func main() {
	initDB() // Инициализируем БД при старте сервера
	defer db.Close()

	http.HandleFunc("/", helloHandler)
	
	fmt.Println("Сервер запущен на порту 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
