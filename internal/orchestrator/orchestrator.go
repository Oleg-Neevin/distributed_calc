package orchestrator

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Oleg-Neevin/distributed_calc/pkg"
)

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Expression struct {
	ID     int     `json:"id"`
	Expr   string  `json:"expression"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

var (
	taskQueue     = make(chan Task, 100)
	chTaskResults = make(map[int]chan float64)
	expressions   = make(map[int]*Expression)
	mu            sync.Mutex
	expressionID  int
)

func RunOrchestrator() {
	http.HandleFunc("/api/v1/calculate", handleCalculate)
	http.HandleFunc("/api/v1/expressions", handleExpressions)
	http.HandleFunc("/api/v1/expressions/", handleExpressionByID)
	http.HandleFunc("/internal/task", handleTask)
	http.HandleFunc("/internal/task/result", handleTaskResult)

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expr string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	expressionID++
	expressions[expressionID] = &Expression{ID: expressionID, Expr: req.Expr, Status: "processing"}
	chTaskResults[expressionID] = make(chan float64, 1)
	mu.Unlock()

	go parseExpression(expressionID, req.Expr)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": expressionID})
}

func parseExpression(id int, expression string) {
	var operations []rune
	var numbers []float64

	// Анализируем вводимые данные:

	for i := 0; i < len(expression); i++ {

		// числа и арифметические знаки записываем в списки
		if expression[i] == '*' || expression[i] == '/' || expression[i] == '+' || expression[i] == '-' {
			operations = append(operations, []rune(expression)[i])

		} else if (expression[i] - '0') <= 9 {
			numbers = append(numbers, float64(expression[i]-'0'))
		}
	}

	// проверяем количество чисел и знаков
	if len(numbers) != len(operations)+1 {
		mu.Lock()
		expressions[id] = &Expression{ID: id, Expr: expression, Status: "error", Result: 0}
		mu.Unlock()
		return
	}

	// вычисляем преоритетные операции
	for i := 0; i < len(operations); i++ {
		switch operations[i] {
		case '*':
			addTask(id, "*", numbers[i], numbers[i+1])

			res := <-chTaskResults[id]
			numbers = append(append(numbers[:i], res), numbers[i+2:]...)
			operations = append(operations[:i], operations[i+1:]...)
			i--

		case '/':
			if numbers[i+1] == 0 {
				mu.Lock()
				expressions[id] = &Expression{ID: id, Expr: expression, Status: "error", Result: 0}
				mu.Unlock()
				return
			}
			addTask(id, "/", numbers[i], numbers[i+1])

			res := <-chTaskResults[id]
			numbers = append(append(numbers[:i], res), numbers[i+2:]...)
			operations = append(operations[:i], operations[i+1:]...)
			i--
		}
	}

	// вычисляем менее преоритетные операции
	for i := 0; i < len(operations); i++ {
		switch operations[i] {
		case '+':
			addTask(id, "+", numbers[i], numbers[i+1])

			res := <-chTaskResults[id]
			numbers = append(append(numbers[:i], res), numbers[i+2:]...)
			operations = append(operations[:i], operations[i+1:]...)
			i--

		case '-':
			addTask(id, "-", numbers[i], numbers[i+1])

			res := <-chTaskResults[id]
			numbers = append(append(numbers[:i], res), numbers[i+2:]...)
			operations = append(operations[:i], operations[i+1:]...)
			i--
		}
	}

	mu.Lock()
	expressions[id] = &Expression{ID: id, Expr: expression, Status: "completed", Result: numbers[0]}
	mu.Unlock()
}

func addTask(id int, op string, arg1, arg2 float64) {
	opTime := getOperationTime(op)
	taskQueue <- Task{ID: id, Arg1: arg1, Arg2: arg2, Operation: op, OperationTime: opTime}
}

func getOperationTime(op string) int {
	switch op {
	case "+":
		return pkg.GetEnvInt("TIME_ADDITION_MS", 100)
	case "-":
		return pkg.GetEnvInt("TIME_SUBTRACTION_MS", 100)
	case "*":
		return pkg.GetEnvInt("TIME_MULTIPLICATIONS_MS", 200)
	case "/":
		return pkg.GetEnvInt("TIME_DIVISIONS_MS", 300)
	}
	return 100
}

func handleExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	var expList []Expression
	for _, exp := range expressions {
		expList = append(expList, *exp)
	}
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expList})
}

func handleExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/api/v1/expressions/"):]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	exp, found := expressions[id]
	mu.Unlock()

	if !found {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": exp})
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		select {
		case task := <-taskQueue:
			json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
		default:
			http.Error(w, "No tasks", http.StatusNotFound)
		}
		return
	} else if r.Method == http.MethodPost {
		var res struct {
			ID     int     `json:"id"`
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		mu.Lock()
		ch, exists := chTaskResults[res.ID]
		mu.Unlock()

		if !exists {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		ch <- res.Result
		w.WriteHeader(http.StatusOK)
		return
	}
}

func handleTaskResult(w http.ResponseWriter, r *http.Request) {
	var res struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	chTaskResults[res.ID] <- res.Result
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}
