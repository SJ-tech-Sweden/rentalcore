package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
)

func TestCreateManualRentalEntryValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRentalEquipmentHandler(nil)
	r.POST("/api/v1/rental-equipment/manual-entry", h.CreateManualRentalEntry)

	// send empty payload
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rental-equipment/manual-entry", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body["error"] != "validation failed" {
		t.Fatalf("expected validation failed error, got %v", body["error"])
	}
	if _, ok := body["fields"]; !ok {
		t.Fatalf("expected fields map in response")
	}
}

func TestAddRentalToJobValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRentalEquipmentHandler(nil)
	r.POST("/api/v1/rental-equipment/add-to-job", h.AddRentalToJob)

	payload := models.AddRentalToJobRequest{}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rental-equipment/add-to-job", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body["error"] != "validation failed" {
		t.Fatalf("expected validation failed error, got %v", body["error"])
	}
}
