package services

import (
	"encoding/json"
	"testing"

	"go-barcode-webapp/internal/models"
)

// ---------------------------------------------------------------------------
// TwentyService – GetConfig / SaveConfig tests
// ---------------------------------------------------------------------------

func TestTwentyService_GetConfig_Defaults(t *testing.T) {
	db := newTestDB(t)
	svc := NewTwentyService(db)

	cfg := svc.GetConfig()
	if cfg.Enabled {
		t.Error("expected Enabled=false for empty config")
	}
	if cfg.APIURL != "" {
		t.Errorf("expected empty APIURL, got %q", cfg.APIURL)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
}

func TestTwentyService_SaveAndGetConfig(t *testing.T) {
	db := newTestDB(t)
	svc := NewTwentyService(db)

	want := TwentyConfig{
		Enabled: true,
		APIURL:  "https://crm.example.com",
		APIKey:  "secret-api-key",
	}
	if err := svc.SaveConfig(want); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got := svc.GetConfig()
	if got.Enabled != want.Enabled {
		t.Errorf("Enabled: got %v, want %v", got.Enabled, want.Enabled)
	}
	if got.APIURL != want.APIURL {
		t.Errorf("APIURL: got %q, want %q", got.APIURL, want.APIURL)
	}
	if got.APIKey != want.APIKey {
		t.Errorf("APIKey: got %q, want %q", got.APIKey, want.APIKey)
	}
}

func TestTwentyService_SaveConfig_Upsert(t *testing.T) {
	db := newTestDB(t)
	svc := NewTwentyService(db)

	// First save
	if err := svc.SaveConfig(TwentyConfig{Enabled: true, APIURL: "https://a.example.com", APIKey: "key1"}); err != nil {
		t.Fatalf("first SaveConfig: %v", err)
	}
	// Second save (update)
	if err := svc.SaveConfig(TwentyConfig{Enabled: false, APIURL: "https://b.example.com", APIKey: "key2"}); err != nil {
		t.Fatalf("second SaveConfig: %v", err)
	}

	got := svc.GetConfig()
	if got.Enabled {
		t.Error("expected Enabled=false after second save")
	}
	if got.APIURL != "https://b.example.com" {
		t.Errorf("APIURL: got %q, want %q", got.APIURL, "https://b.example.com")
	}
	if got.APIKey != "key2" {
		t.Errorf("APIKey: got %q, want %q", got.APIKey, "key2")
	}

	// Confirm only one row per key (no duplicate rows)
	var count int64
	db.Model(&models.AppSetting{}).Where("key = ?", TwentyEnabledKey).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row for %q, got %d", TwentyEnabledKey, count)
	}
}

func TestTwentyService_SaveConfig_TrailingSlash(t *testing.T) {
	db := newTestDB(t)
	svc := NewTwentyService(db)

	if err := svc.SaveConfig(TwentyConfig{Enabled: true, APIURL: "https://a.example.com/", APIKey: "k"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	got := svc.GetConfig()
	if got.APIURL != "https://a.example.com" {
		t.Errorf("expected trailing slash stripped, got %q", got.APIURL)
	}
}

// ---------------------------------------------------------------------------
// Helper / pure-function tests
// ---------------------------------------------------------------------------

func TestMapJobStatusToStage(t *testing.T) {
	cases := []struct {
		status string
		want   string
	}{
		{"New", "NEW"},
		{"Open", "NEW"},
		{"In Progress", "IN_PROGRESS"},
		{"Active", "IN_PROGRESS"},
		{"Completed", "CLOSED_WON"},
		{"Done", "CLOSED_WON"},
		{"Won", "CLOSED_WON"},
		{"Lost", "CLOSED_LOST"},
		{"Cancelled", "CLOSED_LOST"},
		{"something else", "NEW"},
	}
	for _, tc := range cases {
		if got := mapJobStatusToStage(tc.status); got != tc.want {
			t.Errorf("mapJobStatusToStage(%q): got %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestExtractID(t *testing.T) {
	raw := map[string]json.RawMessage{
		"createCompany": []byte(`{"id":"abc-123","name":"Acme"}`),
	}
	id := extractID(raw, "createCompany")
	if id != "abc-123" {
		t.Errorf("extractID: got %q, want %q", id, "abc-123")
	}

	id2 := extractID(raw, "notPresent")
	if id2 != "" {
		t.Errorf("extractID missing key: got %q, want empty", id2)
	}
}

func TestBuildAddress(t *testing.T) {
	street := "Main St"
	house := "12"
	city := "Berlin"
	state := "BE"
	zip := "10115"
	country := "Germany"

	c := &models.Customer{
		Street:       &street,
		HouseNumber:  &house,
		City:         &city,
		FederalState: &state,
		ZIP:          &zip,
		Country:      &country,
	}

	addr := buildAddress(c)
	if addr["addressStreet1"] != "Main St 12" {
		t.Errorf("street1: got %q, want %q", addr["addressStreet1"], "Main St 12")
	}
	if addr["addressCity"] != "Berlin" {
		t.Errorf("city: got %q", addr["addressCity"])
	}
	if addr["addressCountry"] != "Germany" {
		t.Errorf("country: got %q", addr["addressCountry"])
	}
}

func TestTwentyService_TestConnection_NoConfig(t *testing.T) {
	db := newTestDB(t)
	svc := NewTwentyService(db)

	err := svc.TestConnection()
	if err == nil {
		t.Error("expected error when no config set")
	}
}
