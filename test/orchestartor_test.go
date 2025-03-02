package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"CalcWebServiceAsync/internal/application"
)

func TestHandleCalculate(t *testing.T) {
	orch := application.NewOrchestrator()
	server := httptest.NewServer(http.HandlerFunc(orch.HandleCalculate))
	defer server.Close()

	reqBody := `{"expression": "2+3*4"}`
	resp, err := http.Post(server.URL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус 201, получен %d", resp.StatusCode)
	}
	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}
	if res["id"] == nil {
		t.Errorf("Ожидался id в ответе")
	}
}

func TestHandleListAndGetExpression(t *testing.T) {
	orch := application.NewOrchestrator()

	reqBody := `{"expression": "1+2"}`
	reqCalc := httptest.NewRequest("POST", "/api/v1/calculate", strings.NewReader(reqBody))
	reqCalc.Header.Set("Content-Type", "application/json")
	wCalc := httptest.NewRecorder()
	orch.HandleCalculate(wCalc, reqCalc)
	resCalc := wCalc.Result()
	if resCalc.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resCalc.StatusCode)
	}
	var respData map[string]interface{}
	if err := json.NewDecoder(resCalc.Body).Decode(&respData); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}
	exprID := int(respData["id"].(float64))

	reqList := httptest.NewRequest("GET", "/api/v1/expressions", nil)
	wList := httptest.NewRecorder()
	orch.HandleListExpressions(wList, reqList)
	resList := wList.Result()
	if resList.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for list expressions, got %d", resList.StatusCode)
	}
	var listResp struct {
		Expressions []application.Expression `json:"expressions"`
	}
	if err := json.NewDecoder(resList.Body).Decode(&listResp); err != nil {
		t.Fatalf("Error decoding list response: %v", err)
	}
	if len(listResp.Expressions) == 0 {
		t.Errorf("Expected хотя бы одно выражение в списке")
	}

	url := "/api/v1/expressions/" + strconv.Itoa(exprID)
	reqGet := httptest.NewRequest("GET", url, nil)
	wGet := httptest.NewRecorder()
	orch.HandleGetExpression(wGet, reqGet)
	resGet := wGet.Result()
	if resGet.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for get expression, got %d", resGet.StatusCode)
	}
	var getResp struct {
		Expression application.Expression `json:"expression"`
	}
	if err := json.NewDecoder(resGet.Body).Decode(&getResp); err != nil {
		t.Fatalf("Error decoding get response: %v", err)
	}
	if getResp.Expression.ID != exprID {
		t.Errorf("Expected expression id %d, got %d", exprID, getResp.Expression.ID)
	}
}
