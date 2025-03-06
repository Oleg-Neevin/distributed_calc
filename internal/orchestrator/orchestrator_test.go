package orchestrator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleCalculate(t *testing.T) {
	go RunOrchestrator()
	time.Sleep(100 * time.Millisecond)

	reqBody, _ := json.Marshal(map[string]string{
		"expression": "2+3",
	})
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/calculate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var response struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID <= 0 {
		t.Errorf("Expected positive ID, got %d", response.ID)
	}
}

func TestHandleTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task" && r.Method == http.MethodGet {

			task := Task{
				ID:            1,
				Arg1:          2,
				Arg2:          3,
				Operation:     "+",
				OperationTime: 100,
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/internal/task")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response struct {
		Task Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Task.ID != 1 || response.Task.Operation != "+" {
		t.Errorf("Unexpected task data: %+v", response.Task)
	}
}

func TestHandleExpressions(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/expressions" && r.Method == http.MethodGet {

			expressions := []Expression{
				{ID: 1, Expr: "2+3", Status: "completed", Result: 5},
				{ID: 2, Expr: "4*5", Status: "processing"},
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
		} else {
			http.Error(w, "Not found", http.StatusNotFound)

		}
	}))

	defer server.Close()
	resp, err := http.Get(server.URL + "/api/v1/expressions")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response struct {
		Expressions []Expression `json:"expressions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Expressions) != 2 {
		t.Errorf("Expected 2 expressions, got %d", len(response.Expressions))
	}
}
