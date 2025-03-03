package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Computer struct {
	name    string
	client  *http.Client
	baseURL string
}

func NewComputer(name string) *Computer {
	return &Computer{
		name:    name,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "http://localhost:8080",
	}
}

func (c *Computer) getTask() (*pkgModels.Task, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/internal/task", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var taskResp pkgModels.TaskResponse
	err = json.NewDecoder(resp.Body).Decode(&taskResp)
	if err != nil {
		return nil, err
	}

	return taskResp.Task, nil
}

func (c *Computer) processTask(task *pkgModels.Task) (float64, error) {
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return task.Arg1 / task.Arg2, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", task.Operation)
	}
}

func (c *Computer) reportError(taskID string, err error) {
	log.Printf("[%s] Ошибка при обработке задачи %s: %v", c.name, taskID, err)
}

func (c *Computer) sendResult(taskID string, result float64) error {
	reqBody := pkgModels.ResultRequest{
		TaskID:  taskID,
		Result:  result,
		Updated: time.Now(),
		Status:  "completed",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/task", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// Run запускает процесс обработки задач вычислителя
func (c *Computer) Run() error {
	for {
		task, err := c.getTask()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		result, err := c.processTask(task)
		if err != nil {
			c.reportError(task.ID, err)
			continue
		}

		if err := c.sendResult(task.ID, result); err != nil {
			log.Printf("[%s] Ошибка отправки результата: %v", c.name, err)
		}
	}
}

func main() {
	cfg := pkgConfig.LoadConfig()
	var wg sync.WaitGroup

	for i := 0; i < cfg.ComputingPower; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			computer := NewComputer(fmt.Sprintf("computer-%d", id))
			if err := computer.Run(); err != nil {
				log.Printf("Вычислитель %d остановлен с ошибкой: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}
