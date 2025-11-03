package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBuildWarehouseProductsURLWithEnv(t *testing.T) {
	const domain = "warehouse.example.com"
	if err := os.Setenv("WAREHOUSECORE_DOMAIN", domain); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv("WAREHOUSECORE_DOMAIN")
	})

	req := httptest.NewRequest(http.MethodGet, "https://rental.example.com/products", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	got := buildWarehouseProductsURL(req)
	want := "https://" + domain + "/admin/products"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestBuildWarehouseProductsURLFallback(t *testing.T) {
	_ = os.Unsetenv("WAREHOUSECORE_DOMAIN")

	req := httptest.NewRequest(http.MethodGet, "https://rent.example.com/products", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	got := buildWarehouseProductsURL(req)
	want := "https://warehouse.example.com/admin/products"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestBuildWarehouseDevicesURLWithEnv(t *testing.T) {
	const domain = "warehouse.example.com"
	if err := os.Setenv("WAREHOUSECORE_DOMAIN", domain); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv("WAREHOUSECORE_DOMAIN")
	})

	req := httptest.NewRequest(http.MethodGet, "https://rental.example.com/devices", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	got := buildWarehouseDevicesURL(req)
	want := "https://" + domain + "/admin/devices"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestBuildWarehouseDevicesURLFallback(t *testing.T) {
	_ = os.Unsetenv("WAREHOUSECORE_DOMAIN")

	req := httptest.NewRequest(http.MethodGet, "https://rent.example.com/devices", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	got := buildWarehouseDevicesURL(req)
	want := "https://warehouse.example.com/admin/devices"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestBuildWarehouseCablesURLWithEnv(t *testing.T) {
	const domain = "warehouse.example.com"
	if err := os.Setenv("WAREHOUSECORE_DOMAIN", domain); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv("WAREHOUSECORE_DOMAIN")
	})

	req := httptest.NewRequest(http.MethodGet, "https://rental.example.com/cables", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	got := buildWarehouseCablesURL(req)
	want := "https://" + domain + "/admin/cables"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestBuildWarehouseCablesURLFallback(t *testing.T) {
	_ = os.Unsetenv("WAREHOUSECORE_DOMAIN")

	req := httptest.NewRequest(http.MethodGet, "https://rent.example.com/cables", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	got := buildWarehouseCablesURL(req)
	want := "https://warehouse.example.com/admin/cables"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestBuildWarehouseCablesURLWithPort(t *testing.T) {
	_ = os.Unsetenv("WAREHOUSECORE_DOMAIN")

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8081/cables", nil)

	got := buildWarehouseCablesURL(req)
	want := "http://localhost:8082/admin/cables"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
