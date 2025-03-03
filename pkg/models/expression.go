package models

import (
	"time"
)

// Expression представляет арифметическое выражение
type Expression struct {
	ID        string    `json:"id"`
	Text      string    `json:"expression"`
	Status    string    `json:"status"`
	Result    *float64  `json:"result,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Task представляет отдельную математическую операцию
type Task struct {
	ID            string    `json:"id"`
	ExpressionID  string    `json:"expression_id"`
	Arg1          float64   `json:"arg1"`
	Arg2          float64   `json:"arg2"`
	Operation     string    `json:"operation"`
	Status        string    `json:"status"`
	Result        *float64  `json:"result,omitempty"`
	OperationTime int       `json:"operation_time"` // время выполнения в мс
	CreatedAt     time.Time `json:"created_at"`
}

// TaskResponse структура для ответа агенту при получении задачи
type TaskResponse struct {
	Task *Task `json:"task"`
}

// ResultRequest структура для приема результата от агента
type ResultRequest struct {
	TaskID  string    `json:"id"`
	Result  float64   `json:"result"`
	Status  string    `json:"status"`
	Updated time.Time `json:"updated_at"`
}
