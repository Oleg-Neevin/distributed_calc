package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/Oleg-Neevin/distributed_calc/internal/agent"
	"github.com/Oleg-Neevin/distributed_calc/internal/orchestrator"
)

func TestEndToEnd(t *testing.T) {
	go orchestrator.RunOrchestrator()
	go agent.StartAgent()

	time.Sleep(100 * time.Millisecond)
	expressionRequest := map[string]string{
		"expression": "2+3*4",
	}

	reqBody, _ := json.Marshal(expressionRequest)
	resp, err := http.Post("http://localhost:8080/api/v1/calculate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to send expression: %v", err)
	}

	var createResponse struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	exprID := createResponse.ID
	var status string
	var result float64

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		time.Sleep(200 * time.Millisecond)

		resp, err := http.Get("http://localhost:8080/api/v1/expressions/" + strconv.Itoa(exprID))
		if err != nil {
			t.Fatalf("Failed to get expression status: %v", err)
		}

		var expressionResponse struct {
			Expression struct {
				ID     int     `json:"id"`
				Expr   string  `json:"expression"`
				Status string  `json:"status"`
				Result float64 `json:"result"`
			} `json:"expression"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&expressionResponse); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		status = expressionResponse.Expression.Status
		result = expressionResponse.Expression.Result

		if status == "completed" || status == "error" {
			break
		}
	}
	if status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", status)
	}

	expectedResult := 14.0 // 2+3*4 = 2+12 = 14
	if result != expectedResult {
		t.Errorf("Expected result %f, got %f", expectedResult, result)
	}
}

func TestPerformanceMultipleExpressions(t *testing.T) {
	go orchestrator.RunOrchestrator()
	go agent.StartAgent()

	time.Sleep(100 * time.Millisecond)

	numRequests := 5

	results := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			expression := ""
			switch index % 4 {
			case 0:
				expression = "2+3*4"
			case 1:
				expression = "10/2+5"
			case 2:
				expression = "8-3+2*4"
			case 3:
				expression = "9*2-3/3"
			}

			reqBody, _ := json.Marshal(map[string]string{"expression": expression})
			resp, err := http.Post("http://localhost:8080/api/v1/calculate", "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Logf("Failed to send expression: %v", err)
				results <- false
				return
			}

			var createResponse struct {
				ID int `json:"id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&createResponse); err != nil {
				t.Logf("Failed to decode response: %v", err)
				resp.Body.Close()
				results <- false
				return
			}
			resp.Body.Close()

			maxRetries := 10
			success := false

			for j := 0; j < maxRetries; j++ {
				time.Sleep(200 * time.Millisecond)

				resp, err := http.Get("http://localhost:8080/api/v1/expressions/" + strconv.Itoa(createResponse.ID))
				if err != nil {
					continue
				}

				var expressionResponse struct {
					Expression struct {
						Status string `json:"status"`
					} `json:"expression"`
				}

				if err := json.NewDecoder(resp.Body).Decode(&expressionResponse); err != nil {
					resp.Body.Close()
					continue
				}
				resp.Body.Close()

				if expressionResponse.Expression.Status == "completed" {
					success = true
					break
				}
			}

			results <- success
		}(i)
	}

	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-results {
			successCount++
		}
	}

	if successCount != numRequests {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}
}
