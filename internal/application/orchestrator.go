package application

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"CalcWebServiceAsync/pkg/calculation"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
)

type Expression struct {
	ID         int      `json:"id"`
	Expression string   `json:"expression"`
	Status     string   `json:"status"`
	Result     float64  `json:"result,omitempty"`
	AST        *ASTNode `json:"-"`
}

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Orchestrator struct {
	expressions map[int]*Expression
	tasksQueue  []*Task
	nextExprID  int
	nextTaskID  int
	mu          sync.Mutex
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		expressions: make(map[int]*Expression),
		tasksQueue:  make([]*Task, 0),
		nextExprID:  1,
		nextTaskID:  1,
	}
}

func (o *Orchestrator) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Expression string `json:"expression"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusUnprocessableEntity)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusUnprocessableEntity)
		return
	}
	ast, err := ParseExpression(req.Expression)
	if err != nil {
		http.Error(w, "Invalid Expression: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	o.mu.Lock()
	exprID := o.nextExprID
	o.nextExprID++
	expression := &Expression{
		ID:         exprID,
		Expression: req.Expression,
		Status:     StatusPending,
		AST:        ast,
	}
	o.expressions[exprID] = expression
	o.generateTasks(expression)
	o.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": exprID,
	})
}

func (o *Orchestrator) HandleListExpressions(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	defer o.mu.Unlock()
	exprs := make([]*Expression, 0, len(o.expressions))
	for _, expr := range o.expressions {
		exprs = append(exprs, expr)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": exprs,
	})
}

func (o *Orchestrator) HandleGetExpression(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	idStr := parts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	expr, exists := o.expressions[id]
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expression": expr,
	})
}

func (o *Orchestrator) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.tasksQueue) == 0 {
		http.Error(w, "No Task", http.StatusNotFound)
		return
	}
	task := o.tasksQueue[0]
	o.tasksQueue = o.tasksQueue[1:]
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task": task,
	})
}

func (o *Orchestrator) HandlePostTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusUnprocessableEntity)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusUnprocessableEntity)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	found := false
	for _, expr := range o.expressions {
		if updateASTWithResult(expr.AST, req.ID, req.Result) {
			found = true
			o.generateTasks(expr)
			if isASTComplete(expr.AST) {
				expr.Status = StatusCompleted
				expr.Result = expr.AST.Value
			} else {
				expr.Status = StatusRunning
			}
			break
		}
	}
	if !found {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "result accepted",
	})
}

func (o *Orchestrator) generateTasks(expr *Expression) {
	var traverse func(node *ASTNode)
	traverse = func(node *ASTNode) {
		if node == nil {
			return
		}
		if node.Operator != "" {
			if node.Left != nil && node.Right != nil &&
				node.Left.Operator == "" && node.Right.Operator == "" &&
				!node.Evaluated {
				task := &Task{
					ID:            o.nextTaskID,
					Arg1:          node.Left.Value,
					Arg2:          node.Right.Value,
					Operation:     node.Operator,
					OperationTime: calculation.GetOperationTime(node.Operator),
				}
				o.nextTaskID++
				node.TaskID = task.ID
				o.tasksQueue = append(o.tasksQueue, task)
			}
			traverse(node.Left)
			traverse(node.Right)
		}
	}
	traverse(expr.AST)
}

func updateASTWithResult(node *ASTNode, taskID int, result float64) bool {
	if node == nil {
		return false
	}
	if node.TaskID == taskID {
		node.Value = result
		node.Operator = ""
		node.Left = nil
		node.Right = nil
		node.Evaluated = true
		return true
	}
	if updateASTWithResult(node.Left, taskID, result) {
		return true
	}
	if updateASTWithResult(node.Right, taskID, result) {
		return true
	}
	return false
}

func isASTComplete(node *ASTNode) bool {
	return node != nil && node.Operator == ""
}
