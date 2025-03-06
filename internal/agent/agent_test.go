package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorker(t *testing.T) {
	taskServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"task": Task{
					ID:            1,
					Arg1:          2,
					Arg2:          3,
					Operation:     "+",
					OperationTime: 1,
				},
			})
		} else if r.URL.Path == "/internal/task" && r.Method == http.MethodPost {
			var result struct {
				ID     int     `json:"id"`
				Result float64 `json:"result"`
			}
			if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			if result.ID != 1 || result.Result != 5 {
				t.Errorf("Expected ID=1, result=5, got ID=%d, result=%f", result.ID, result.Result)
			}
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))

	resp, err := http.Get(taskServer.URL + "/internal/task")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	var taskResponse struct {
		Task Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&taskResponse); err != nil {
		t.Fatalf("Failed to decode task: %v", err)
	}
	resp.Body.Close()

	if taskResponse.Task.ID != 1 || taskResponse.Task.Operation != "+" {
		t.Errorf("Unexpected task: %+v", taskResponse.Task)
	}

	result := map[string]interface{}{
		"id":     1,
		"result": 5.0,
	}

	resultJSON, _ := json.Marshal(result)
	resp, err = http.Post(taskServer.URL+"/internal/task", "application/json", bytes.NewBuffer(resultJSON))
	if err != nil {
		t.Fatalf("Failed to post result: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
