package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *repository.Database {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed open sqlite: %v", err)
	}
	// migrate required models
	if err := db.AutoMigrate(&models.Customer{}, &models.Status{}, &models.Job{}, &models.RentalEquipment{}, &models.JobRentalEquipment{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return &repository.Database{db}
}

func TestManualCreateAndAddToJobIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repoDB := setupTestDB(t)

	// prepare minimal job
	gdb := repoDB.DB
	cust := models.Customer{}
	if err := gdb.Create(&cust).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	stat := models.Status{Status: "New"}
	if err := gdb.Create(&stat).Error; err != nil {
		t.Fatalf("create status: %v", err)
	}
	job := models.Job{CustomerID: cust.CustomerID, StatusID: stat.StatusID}
	if err := gdb.Create(&job).Error; err != nil {
		t.Fatalf("create job: %v", err)
	}

	rentalRepo := repository.NewRentalEquipmentRepository(repoDB)
	handler := NewRentalEquipmentHandler(rentalRepo)

	r := gin.New()
	r.POST("/api/v1/rental-equipment/manual-entry", handler.CreateManualRentalEntry)

	payload := models.ManualRentalEntryRequest{
		JobID:        job.JobID,
		ProductName:  "Test Device",
		SupplierName: "Test Supplier",
		RentalPrice:  12.5,
		Quantity:     2,
		DaysUsed:     3,
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rental-equipment/manual-entry", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d body=%s", w.Code, w.Body.String())
	}

	// verify created rental equipment and job rental
	var equipments []models.RentalEquipment
	if err := gdb.Find(&equipments).Error; err != nil {
		t.Fatalf("query equipments: %v", err)
	}
	if len(equipments) != 1 {
		t.Fatalf("expected 1 equipment, got %d", len(equipments))
	}
	var jobRentals []models.JobRentalEquipment
	if err := gdb.Where("job_id = ?", job.JobID).Find(&jobRentals).Error; err != nil {
		t.Fatalf("query job rentals: %v", err)
	}
	if len(jobRentals) != 1 {
		t.Fatalf("expected 1 job rental, got %d", len(jobRentals))
	}
}
