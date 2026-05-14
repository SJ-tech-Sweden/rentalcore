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

func newTestProductDB(t *testing.T) *Database {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// AutoMigrate related tables used by Product preloads to avoid "no such table" errors
	if err := db.AutoMigrate(&models.Brand{}, &models.Category{}, &models.Subcategory{}, &models.Subbiercategory{}, &models.Manufacturer{}, &models.CountType{}, &models.Product{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return &Database{DB: db}
}

func seedProduct(t *testing.T, db *Database, id uint, name string) {
	t.Helper()
	p := models.Product{ProductID: id, Name: name}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
}

func TestProductRepository_GetByID_DefaultDB(t *testing.T) {
	db := newTestProductDB(t)
	seedProduct(t, db, 1, "DB Lamp")

	repo := NewProductRepository(db)
	got, err := repo.GetByID(1)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got == nil || got.Name != "DB Lamp" {
		t.Fatalf("unexpected product: %#v", got)
	}
}

func TestProductRepository_GetByID_APIMode(t *testing.T) {
	// Start an httptest server that mimics WarehouseCore product endpoints
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// return a product for any /service/products/{id}
		if r.URL.Path == "/service/products/42" {
			json.NewEncoder(w).Encode(map[string]interface{}{"id": 42, "name": "API Speaker", "price": 12.5})
			return
		}
		// list
		if r.URL.Path == "/service/products" {
			json.NewEncoder(w).Encode([]map[string]interface{}{{"id": 42, "name": "API Speaker"}})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := newTestProductDB(t)
	// no DB seed for id 42; API should return value
	repo := NewProductRepository(db)
	wc := warehousecore.NewClientWithConfig(srv.URL, "")
	repo.WithWarehouseCoreClient(wc, true)

	got, err := repo.GetByID(42)
	if err != nil {
		t.Fatalf("GetByID API error: %v", err)
	}
	if got == nil || got.Name == "" {
		t.Fatalf("expected API product, got: %#v", got)
	}
	if got.PricePerUnit == nil || *got.PricePerUnit != 12.5 {
		t.Fatalf("expected API product price, got: %#v", got)
	}
}

func TestProductRepository_List_APIMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/service/products" {
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"id": 7, "name": "API Cable", "price": 10.5},
				{"id": 8, "name": "API Cable Pro", "price": 20.25},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := newTestProductDB(t)
	repo := NewProductRepository(db)
	wc := warehousecore.NewClientWithConfig(srv.URL, "")
	repo.WithWarehouseCoreClient(wc, true)

	list, err := repo.List(nil)
	if err != nil {
		t.Fatalf("List API error: %v", err)
	}
	if len(list) != 2 || list[0].Name == "" || list[1].Name == "" {
		t.Fatalf("unexpected list result: %#v", list)
	}
	if list[0].PricePerUnit == nil || list[1].PricePerUnit == nil {
		t.Fatalf("expected non-nil price pointers: %#v", list)
	}
	if *list[0].PricePerUnit != 10.5 || *list[1].PricePerUnit != 20.25 {
		t.Fatalf("unexpected prices in API list: %#v", list)
	}
}

func TestProductRepository_List_APIMode_NonTransientErrorDoesNotFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/service/products" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := newTestProductDB(t)
	seedProduct(t, db, 1, "DB Product Should Not Be Returned")

	repo := NewProductRepository(db)
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

func TestProductRepository_List_APIMode_TransientServerErrorFallsBackToDB(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/service/products" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := newTestProductDB(t)
	seedProduct(t, db, 1, "DB Product Fallback")

	repo := NewProductRepository(db)
	wc := warehousecore.NewClientWithConfig(srv.URL, "")
	repo.WithWarehouseCoreClient(wc, true)

	list, err := repo.List(nil)
	if err != nil {
		t.Fatalf("expected DB fallback on transient API error, got error: %v", err)
	}
	if len(list) == 0 || list[0].Name != "DB Product Fallback" {
		t.Fatalf("expected DB fallback results on transient API error, got: %#v", list)
	}
}
