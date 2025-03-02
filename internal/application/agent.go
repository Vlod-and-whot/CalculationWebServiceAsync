package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"CalcWebServiceAsync/pkg/calculation"
)

type Agent struct {
	orchestratorURL string
	client          *http.Client
}

func NewAgent(orchestratorURL string) *Agent {
	return &Agent{
		orchestratorURL: orchestratorURL,
		client:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *Agent) FetchTask() (*Task, error) {
	resp, err := a.client.Get(a.orchestratorURL + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch task")
	}
	var res struct {
		Task *Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Task, nil
}

func (a *Agent) ExecuteTask(task *Task) (float64, error) {
	return calculation.Compute(task.Arg1, task.Arg2, task.Operation, task.OperationTime)
}

func (a *Agent) SubmitResult(taskID int, result float64) error {
	payload := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := a.client.Post(a.orchestratorURL+"/internal/task", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to submit result")
	}
	return nil
}
