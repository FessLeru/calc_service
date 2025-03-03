// cmd/orchestrator/main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	pkgMiddleware "calc_system/pkg/middleware"
	pkgModels "calc_system/pkg/models"
)

type Orchestrator struct {
	taskQueue   chan pkgModels.Task
	expressions map[string]pkgModels.Expression
	tasks       map[string]pkgModels.Task
	mu          sync.RWMutex
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		taskQueue:   make(chan pkgModels.Task, 100),
		expressions: make(map[string]pkgModels.Expression),
		tasks:       make(map[string]pkgModels.Task),
	}
}

func (o *Orchestrator) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expr := &pkgModels.Expression{
		ID:        generateUUID(),
		Text:      req.Expression,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	o.expressions[expr.ID] = *expr
	o.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{ID: expr.ID})
}

func (o *Orchestrator) HandleExpressionsList(w http.ResponseWriter, r *http.Request) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	expressions := make([]pkgModels.Expression, 0, len(o.expressions))
	for _, expr := range o.expressions {
		expressions = append(expressions, expr)
	}

	json.NewEncoder(w).Encode(struct {
		Expressions []pkgModels.Expression `json:"expressions"`
	}{Expressions: expressions})
}

func (o *Orchestrator) HandleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		o.handleGetTask(w, r)
		return
	}

	var result pkgModels.ResultRequest
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	o.mu.Lock()
	task, ok := o.tasks[result.TaskID]
	if !ok {
		o.mu.Unlock()
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	task.Status = result.Status
	task.Result = &result.Result
	task.UpdatedAt = result.Updated
	o.tasks[result.TaskID] = task

	o.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func main() {
	orchestrator := NewOrchestrator()

	router := http.NewServeMux()
	router.HandleFunc("/api/v1/calculate", orchestrator.HandleCalculate)
	router.HandleFunc("/api/v1/expressions", orchestrator.HandleExpressionsList)
	router.HandleFunc("/internal/task", orchestrator.HandleTask)

	middlewareChain := pkgMiddleware.RateLimitMiddleware(
		pkgMiddleware.RecoveryMiddleware(
			pkgMiddleware.LoggingMiddleware(router)))

	log.Println("Оркестратор запущен на порту :8080")
	if err := http.ListenAndServe(":8080", middlewareChain); err != nil {
		log.Fatal(err)
	}
}
