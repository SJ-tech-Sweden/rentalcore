package warehouseclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-barcode-webapp/internal/config"
)

// Client is a minimal HTTP client for calling WarehouseCore service APIs.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient constructs a client from RentalCore config (reads WarehouseCore section).
func NewClient(cfg *config.Config) *Client {
	base := strings.TrimRight(cfg.WarehouseCore.BaseURL, "/")
	return &Client{
		baseURL: base,
		apiKey:  cfg.WarehouseCore.APIKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Device represents the subset of device fields used by RentalCore.
type Device struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	CableID  *int   `json:"cable_id,omitempty"`
}

// Cable represents cable metadata returned by WarehouseCore.
type Cable struct {
	CableID    int      `json:"cable_id"`
	Name       string   `json:"name"`
	Connector1 *int     `json:"connector1,omitempty"`
	Connector2 *int     `json:"connector2,omitempty"`
	Typ        *int     `json:"typ,omitempty"`
	Length     *float64 `json:"length,omitempty"`
	Mm2        *float64 `json:"mm2,omitempty"`
}

// Product is a minimal product representation (extend as needed).
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	SKU   string  `json:"sku,omitempty"`
	Price float64 `json:"price,omitempty"`
}

// doGet performs a GET request to the WarehouseCore service and decodes JSON into v.
func (c *Client) doGet(ctx context.Context, path string, v interface{}) error {
	if c.baseURL == "" {
		return errors.New("warehousecore base URL not configured")
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", url)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("warehousecore error: status=%d body=%s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(v)
}

// GetDevice fetches device metadata by device ID.
func (c *Client) GetDevice(ctx context.Context, deviceID string) (*Device, error) {
	var d Device
	path := fmt.Sprintf("/devices/%s", deviceID)
	if err := c.doGet(ctx, path, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// GetCable fetches cable metadata by cable ID.
func (c *Client) GetCable(ctx context.Context, cableID int) (*Cable, error) {
	var cb Cable
	path := fmt.Sprintf("/admin/cables/%d", cableID)
	if err := c.doGet(ctx, path, &cb); err != nil {
		return nil, err
	}
	return &cb, nil
}

// GetProduct fetches product metadata by product ID (endpoint may vary in WarehouseCore).
func (c *Client) GetProduct(ctx context.Context, productID int) (*Product, error) {
	var p Product
	path := fmt.Sprintf("/service/products/%d", productID)
	if err := c.doGet(ctx, path, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// SearchCables allows a simple search call to /admin/cables?search=...
func (c *Client) SearchCables(ctx context.Context, query string) ([]Cable, error) {
	var items []Cable
	path := fmt.Sprintf("/admin/cables?search=%s", urlQueryEscape(query))
	if err := c.doGet(ctx, path, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// minimal URL query escaper – keeps stdlib import surface small
func urlQueryEscape(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), " ", "%20")
}
