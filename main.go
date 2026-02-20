package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// Структура для приема и отправки данных
type ScoreRecord struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func initDB() {
	var err error
	dbPath := "/data/snake.db"
	if _, err := os.Stat("/data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
		dbPath = "data/snake.db"
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Ошибка открытия БД:", err)
	}

	// Создаем таблицу рекордов, если её нет
	query := `
	CREATE TABLE IF NOT EXISTS daily_scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		score INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err = db.Exec(query); err != nil {
		log.Fatal("Ошибка создания таблицы:", err)
	}
}

// Встроенный HTML с добавленной логикой рекордов
const snakeHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>Игра Питон (Змейка)</title>
    <style>
        body {
            background-color: #2b2b2b; color: #ffffff;
            font-family: 'Segoe UI', Tahoma, sans-serif;
            margin: 0; padding: 0;
            display: flex; flex-direction: column;
            align-items: center; justify-content: center;
            min-height: 100vh; touch-action: none;
        }
        h1 { margin: 10px 0; font-size: 24px; }
        canvas {
            background-color: #000000;
            box-shadow: 0 0 15px rgba(0, 255, 0, 0.2);
            border: 2px solid #4CAF50;
            max-width: 95vw; max-height: 50vh;
        }
        .controls {
            display: grid; grid-template-columns: 60px 60px 60px;
            gap: 10px; margin-top: 15px;
        }
        .btn {
            width: 60px; height: 60px;
            background: #444; border: 2px solid #555;
            border-radius: 12px; color: white;
            font-size: 24px; display: flex;
            align-items: center; justify-content: center;
            user-select: none; cursor: pointer;
        }
        .btn:active { background: #666; }
        .empty { background: transparent; border: none; }
        
        #leaderboard {
            margin-top: 15px; padding: 10px;
            background: #333; border-radius: 8px;
            width: 80%; max-width: 300px; text-align: center;
        }
        #leaderboard h2 { margin: 0 0 10px 0; font-size: 18px; color: #4CAF50; }
        #leaderboard ol { margin: 0; padding-left: 20px; text-align: left; }
        
        @media (min-width: 768px) { .controls { display: none; } }
    </style>
</head>
<body>
    <h1>Счет: <span id="score">0</span></h1>
    <canvas id="gameCanvas" width="400" height="400"></canvas>

    <div class="controls">
        <div class="empty"></div><div class="btn" id="up">↑</div><div class="empty"></div>
        <div class="btn" id="left">←</div><div class="btn" id="down">↓</div><div class="btn" id="right">→</div>
    </div>

    <div id="leaderboard">
        <h2>Топ за сегодня</h2>
        <ol id="scoreList"><li>Загрузка...</li></ol>
    </div>

    <script>
        const canvas = document.getElementById("gameCanvas");
        const ctx = canvas.getContext("2d");
        const gridSize = 20;
        let score = 0;
        let dx = gridSize, dy = 0;
        let snake = [{x: 200, y: 200}];
        let food = {x: 0, y: 0};
        let isGameOver = false;

        async function loadLeaderboard() {
            try {
                let res = await fetch('/api/scores');
                let scores = await res.json();
                let list = document.getElementById('scoreList');
                list.innerHTML = '';
                if (!scores || scores.length === 0) {
                    list.innerHTML = '<li>Пока нет рекордов</li>';
                    return;
                }
                scores.forEach(s => {
                    let li = document.createElement('li');
                    li.innerText = s.name + " - " + s.score;
                    list.appendChild(li);
                });
            } catch (e) { console.error(e); }
        }

        async function saveScoreAndReload() {
            if (score > 0) {
                let name = prompt("Игра окончена!\nВаш счет: " + score + "\nВведите имя для таблицы рекордов:", "Игрок");
                if (name) {
                    await fetch('/api/score', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({name: name.substring(0, 15), score: score})
                    });
                }
            } else {
                alert("Игра окончена! Вы ничего не съели.");
            }
            document.location.reload();
        }

        function randomFood() {
            food.x = Math.floor(Math.random() * (canvas.width / gridSize)) * gridSize;
            food.y = Math.floor(Math.random() * (canvas.height / gridSize)) * gridSize;
        }

        function main() {
            if (isGameOver) return;
            if (hasGameEnded()) {
                isGameOver = true;
                saveScoreAndReload();
                return;
            }
            setTimeout(() => {
                clearCanvas(); drawFood(); advanceSnake(); drawSnake(); main();
            }, 100);
        }

        function clearCanvas() { ctx.fillStyle = "black"; ctx.fillRect(0, 0, canvas.width, canvas.height); }
        function drawFood() { ctx.fillStyle = "red"; ctx.fillRect(food.x, food.y, gridSize, gridSize); }
        function drawSnake() {
            snake.forEach((part, index) => {
                ctx.fillStyle = index === 0 ? "#4CAF50" : "#8BC34A";
                ctx.strokeStyle = "#1b5e20";
                ctx.fillRect(part.x, part.y, gridSize, gridSize);
                ctx.strokeRect(part.x, part.y, gridSize, gridSize);
            });
        }
        function advanceSnake() {
            const head = {x: snake[0].x + dx, y: snake[0].y + dy};
            snake.unshift(head);
            if (head.x === food.x && head.y === food.y) {
                score += 10; document.getElementById('score').innerText = score; randomFood();
            } else { snake.pop(); }
        }
        function hasGameEnded() {
            for (let i = 4; i < snake.length; i++) {
                if (snake[i].x === snake[0].x && snake[i].y === snake[0].y) return true;
            }
            return snake[0].x < 0 || snake[0].x >= canvas.width || snake[0].y < 0 || snake[0].y >= canvas.height;
        }

        document.addEventListener("keydown", (e) => {
            const LEFT = 37, UP = 38, RIGHT = 39, DOWN = 40;
            const W = 87, A = 65, S = 83, D = 68;
            if ((e.keyCode === LEFT || e.keyCode === A) && dx === 0) { dx = -gridSize; dy = 0; }
            if ((e.keyCode === UP || e.keyCode === W) && dy === 0) { dx = 0; dy = -gridSize; }
            if ((e.keyCode === RIGHT || e.keyCode === D) && dx === 0) { dx = gridSize; dy = 0; }
            if ((e.keyCode === DOWN || e.keyCode === S) && dy === 0) { dx = 0; dy = gridSize; }
        });

        function bindTouch(id, newDx, newDy) {
            document.getElementById(id).addEventListener('touchstart', (e) => {
                e.preventDefault();
                if (newDx !== 0 && dx === 0) { dx = newDx; dy = 0; }
                if (newDy !== 0 && dy === 0) { dx = 0; dy = newDy; }
            }, {passive: false});
        }
        bindTouch('up', 0, -gridSize); bindTouch('down', 0, gridSize);
        bindTouch('left', -gridSize, 0); bindTouch('right', gridSize, 0);

        loadLeaderboard();
        randomFood();
        main();
    </script>
</body>
</html>`

func gameHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, snakeHTML)
}

// API: Получение рекордов за день
func apiGetScores(w http.ResponseWriter, r *http.Request) {
	// Фильтруем записи по сегодняшней дате (UTC)
	rows, err := db.Query(`
		SELECT name, score 
		FROM daily_scores 
		WHERE date(created_at) = date('now') 
		ORDER BY score DESC LIMIT 5
	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var scores []ScoreRecord
	for rows.Next() {
		var s ScoreRecord
		if err := rows.Scan(&s.Name, &s.Score); err == nil {
			scores = append(scores, s)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// API: Сохранение рекорда
func apiPostScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", 405)
		return
	}

	var s ScoreRecord
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Неверные данные", 400)
		return
	}

	if s.Name != "" && s.Score > 0 {
		_, err := db.Exec("INSERT INTO daily_scores (name, score) VALUES (?, ?)", s.Name, s.Score)
		if err != nil {
			log.Println("Ошибка записи рекорда:", err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	initDB()
	defer db.Close()

	// Роуты
	http.HandleFunc("/", gameHandler)
	http.HandleFunc("/api/scores", apiGetScores)
	http.HandleFunc("/api/score", apiPostScore)

	fmt.Println("Игра 'Питон' с рекордами запущена на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}