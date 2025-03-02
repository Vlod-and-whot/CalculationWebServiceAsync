package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"CalcWebServiceAsync/internal/application"
)

func main() {
	cpStr := os.Getenv("COMPUTING_POWER")
	computingPower := 2
	if cpStr != "" {
		if cp, err := strconv.Atoi(cpStr); err == nil {
			computingPower = cp
		}
	}

	orchestratorAddr := os.Getenv("ORCHESTRATOR_ADDR")
	if orchestratorAddr == "" {
		orchestratorAddr = "http://localhost:8080"
	}

	agent := application.NewAgent(orchestratorAddr)

	var wg sync.WaitGroup
	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				task, err := agent.FetchTask()
				if err != nil {
					log.Printf("Worker %d: ошибка получения задачи: %v", workerID, err)
					time.Sleep(2 * time.Second)
					continue
				}
				if task == nil {
					time.Sleep(2 * time.Second)
					continue
				}
				log.Printf("Worker %d: получила задачу %d: %s", workerID, task.ID, task.Operation)
				result, err := agent.ExecuteTask(task)
				if err != nil {
					log.Printf("Worker %d: ошибка выполнения задачи %d: %v", workerID, task.ID, err)
					continue
				}
				err = agent.SubmitResult(task.ID, result)
				if err != nil {
					log.Printf("Worker %d: ошибка отправки результата задачи %d: %v", workerID, task.ID, err)
				} else {
					log.Printf("Worker %d: успешно обработала задачу %d, результат: %v", workerID, task.ID, result)
				}
			}
		}(i + 1)
	}

	wg.Wait()
}
