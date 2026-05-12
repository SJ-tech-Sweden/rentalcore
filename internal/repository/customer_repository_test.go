package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/services/warehousecore"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestCustomerDB(t *testing.T) *Database {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Customer{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return &Database{DB: db}
}

func seedCustomer(t *testing.T, db *Database, id uint, company string) {
	t.Helper()
	c := models.Customer{CustomerID: id, CompanyName: &company}
	if err := db.Create(&c).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}
}

func TestCustomerRepository_GetByID_DefaultDB(t *testing.T) {
	db := newTestCustomerDB(t)
	seedCustomer(t, db, 1, "DB Co")

	repo := NewCustomerRepository(db)
	got, err := repo.GetByID(1)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got == nil || got.CompanyName == nil || *got.CompanyName != "DB Co" {
		t.Fatalf("unexpected customer: %#v", got)
	}
}

func TestCustomerRepository_GetByID_APIMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/admin/customers/5" {
			json.NewEncoder(w).Encode(map[string]interface{}{"id": 5, "company_name": "API Ltd"})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := newTestCustomerDB(t)
	repo := NewCustomerRepository(db)
	wc := warehousecore.NewClientWithConfig(srv.URL, "")
	repo.WithWarehouseCoreClient(wc, true)

	got, err := repo.GetByID(5)
	if err != nil {
		t.Fatalf("GetByID API error: %v", err)
	}
	if got == nil || got.CompanyName == nil || *got.CompanyName != "API Ltd" {
		t.Fatalf("unexpected API customer: %#v", got)
	}
}

func TestCustomerRepository_List_APIMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/customers" {
			json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 9, "company_name": "API Customer"}})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := newTestCustomerDB(t)
	repo := NewCustomerRepository(db)
	wc := warehousecore.NewClientWithConfig(srv.URL, "")
	repo.WithWarehouseCoreClient(wc, true)

	list, err := repo.List(nil)
	if err != nil {
		t.Fatalf("List API error: %v", err)
	}
	if len(list) == 0 || list[0].CompanyName == nil || *list[0].CompanyName != "API Customer" {
		t.Fatalf("unexpected list result: %#v", list)
	}
}

func TestCustomerRepository_List_APIMode_NonTransientErrorDoesNotFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/customers" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := newTestCustomerDB(t)
	seedCustomer(t, db, 1, "DB Customer Should Not Be Returned")

	repo := NewCustomerRepository(db)
	wc := warehousecore.NewClientWithConfig(srv.URL, "")
	repo.WithWarehouseCoreClient(wc, true)

	list, err := repo.List(nil)
	if err == nil {
		t.Fatalf("expected list error for non-transient API failure, got nil")
	}
	if len(list) != 0 {
		t.Fatalf("expected no DB fallback results on non-transient API error, got: %#v", list)
	}
}
