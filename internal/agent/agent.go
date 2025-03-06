package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Oleg-Neevin/distributed_calc/pkg"
)

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

func StartAgent() {
	computingPower := pkg.GetEnvInt("COMPUTING_POWER", 3)
	for i := 0; i < computingPower; i++ {
		go worker()
	}
}

func worker() {
	for {
		resp, err := http.Get("http://localhost:8080/internal/task")
		if err != nil {
			log.Print(err)
			time.Sleep(1 * time.Second)
			continue
		}
		defer resp.Body.Close()

		var res struct {
			Task Task `json:"task"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			resp.Body.Close()
			continue
		}

		time.Sleep(time.Duration(res.Task.OperationTime) * time.Millisecond)
		result := compute(res.Task.Arg1, res.Task.Arg2, res.Task.Operation)
		postResult(res.Task.ID, result)
	}
}

func compute(arg1, arg2 float64, op string) float64 {
	switch op {
	case "+":
		return arg1 + arg2
	case "-":
		return arg1 - arg2
	case "*":
		return arg1 * arg2
	case "/":
		if arg2 != 0 {
			return arg1 / arg2
		}
	}
	return 0
}

func postResult(id int, result float64) {
	data, _ := json.Marshal(map[string]interface{}{"id": id, "result": result})
	_, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error posting result for task %d: %v", id, err)
	}
}
