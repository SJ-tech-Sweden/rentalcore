package warehousecore

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrCableNotFound is returned by GetCable when WarehouseCore responds with 404.
var ErrCableNotFound = errors.New("cable not found in WarehouseCore")

// RentalEquipmentItem represents a rental equipment item from WarehouseCore
type RentalEquipmentItem struct {
	EquipmentID   uint    `json:"equipment_id"`
	ProductName   string  `json:"product_name"`
	SupplierName  string  `json:"supplier_name"`
	RentalPrice   float64 `json:"rental_price"`
	CustomerPrice float64 `json:"customer_price"`
	Category      string  `json:"category"`
	Subcategory   string  `json:"subcategory"`
	Description   string  `json:"description"`
	IsActive      bool    `json:"is_active"`
}

// Product represents a small product payload returned by WarehouseCore's
// product endpoints. Fields kept minimal to match ProductRepository uses.
type Product struct {
	ID    uint    `json:"id"`
	Name  string  `json:"name"`
	SKU   string  `json:"sku,omitempty"`
	Price float64 `json:"price,omitempty"`
}

// ProductPackage represents package data returned by WarehouseCore.
type ProductPackage struct {
	PackageID   int     `json:"package_id"`
	PackageCode string  `json:"package_code,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty"`
	TotalItems  int     `json:"total_items,omitempty"`
}

type ProductPackageItemDetail struct {
	PackageItemID int    `json:"package_item_id"`
	ProductID     int    `json:"product_id"`
	ProductName   string `json:"product_name"`
	Quantity      int    `json:"quantity"`
}

type ProductPackageDetail struct {
	ProductPackage
	Items []ProductPackageItemDetail `json:"items,omitempty"`
}

// Customer represents a customer payload returned by WarehouseCore.
type Customer struct {
	ID          uint    `json:"id"`
	CompanyName *string `json:"company_name,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	Email       *string `json:"email,omitempty"`
}

// User represents a minimal user payload returned by auth/user endpoints.
type User struct {
	ID        uint    `json:"id"`
	Username  string  `json:"username"`
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// CableSnapshot represents cable metadata fetched from the WarehouseCore API.
// This is a point-in-time copy stored in job_cables.cable_snapshot (JSONB).
type CableSnapshot struct {
	CableID    int      `json:"cableID"`
	Connector1 int      `json:"connector1"`
	Connector2 int      `json:"connector2"`
	Type       int      `json:"typ"`
	Length     float64  `json:"length"`
	MM2        *float64 `json:"mm2,omitempty"`
	Name       *string  `json:"name,omitempty"`
}

// DeviceTreeDevice represents a device item from WarehouseCore /devices/tree.
type DeviceTreeDevice struct {
	DeviceID    string `json:"device_id"`
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
}

// DeviceTreeSubbiercategory mirrors WarehouseCore tree hierarchy.
type DeviceTreeSubbiercategory struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Devices []DeviceTreeDevice `json:"devices"`
}

// DeviceTreeSubcategory mirrors WarehouseCore tree hierarchy.
type DeviceTreeSubcategory struct {
	ID                string                      `json:"id"`
	Name              string                      `json:"name"`
	DirectDevices     []DeviceTreeDevice          `json:"direct_devices"`
	Subbiercategories []DeviceTreeSubbiercategory `json:"subbiercategories"`
}

// DeviceTreeCategory mirrors WarehouseCore tree hierarchy.
type DeviceTreeCategory struct {
	ID            int                     `json:"id"`
	Name          string                  `json:"name"`
	DirectDevices []DeviceTreeDevice      `json:"direct_devices"`
	Subcategories []DeviceTreeSubcategory `json:"subcategories"`
}

// Client is a client for communicating with WarehouseCore API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	mu         sync.RWMutex
	cache      []RentalEquipmentItem
	cacheTime  time.Time
	cacheTTL   time.Duration
}

// NewClient creates a new WarehouseCore client using environment variables.
func NewClient() *Client {
	// Get the WarehouseCore domain from environment variable
	domain := os.Getenv("WAREHOUSECORE_DOMAIN")
	if domain == "" {
		// Fallback for development
		domain = "localhost:8082"
	}

	// Determine protocol based on domain
	protocol := "https"
	if strings.Contains(domain, "localhost") || strings.Contains(domain, "127.0.0.1") {
		protocol = "http"
	}

	baseURL := fmt.Sprintf("%s://%s", protocol, strings.TrimSuffix(domain, "/"))

	return &Client{
		baseURL: baseURL,
		apiKey:  os.Getenv("WAREHOUSECORE_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheTTL: 5 * time.Minute,
	}
}

// NewClientWithURL creates a client with a specific base URL
func NewClientWithURL(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheTTL: 5 * time.Minute,
	}
}

// NewClientWithConfig creates a client with explicit base URL and API key.
func NewClientWithConfig(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheTTL: 5 * time.Minute,
	}
}

// GetBaseURL returns the configured base URL
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// addAuthHeader adds the X-API-Key header when an API key is configured.
func (c *Client) addAuthHeader(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
}

// GetCable fetches cable metadata from WarehouseCore using GET /admin/cables/{id}.
// Returns ErrCableNotFound (wrapped) when the API responds with 404.
func (c *Client) GetCable(id int) (*CableSnapshot, error) {
	url := fmt.Sprintf("%s/admin/cables/%d", c.baseURL, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create cable request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch cable %d: %w", id, err)
	}
	defer func() {
		// Drain any remaining body content to allow keep-alive connection reuse.
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w (id=%d)", ErrCableNotFound, id)
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("WarehouseCore returned %d for cable %d", resp.StatusCode, id)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching cable %d", resp.StatusCode, id)
	}

	var snap CableSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return nil, fmt.Errorf("decode cable %d: %w", id, err)
	}

	return &snap, nil
}

// GetRentalEquipment fetches rental equipment from WarehouseCore
func (c *Client) GetRentalEquipment() ([]RentalEquipmentItem, error) {
	// Check cache first
	c.mu.RLock()
	if len(c.cache) > 0 && time.Since(c.cacheTime) < c.cacheTTL {
		result := make([]RentalEquipmentItem, len(c.cache))
		copy(result, c.cache)
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Fetch from WarehouseCore
	url := fmt.Sprintf("%s/rental-equipment", c.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch rental equipment: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rental equipment API returned status %d", resp.StatusCode)
	}

	var items []RentalEquipmentItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode rental equipment: %w", err)
	}

	normalizeRentalEquipmentIDs(items)

	// Update cache
	c.mu.Lock()
	c.cache = items
	c.cacheTime = time.Now()
	c.mu.Unlock()

	return items, nil
}

// normalizeRentalEquipmentIDs ensures every rental row has a stable non-zero
// equipment ID. Some external feeds return equipment_id=0 which causes
// overwrites in selection maps keyed by equipment ID.
func normalizeRentalEquipmentIDs(items []RentalEquipmentItem) {
	for i := range items {
		if items[i].EquipmentID != 0 {
			continue
		}

		signature := strings.ToLower(strings.TrimSpace(items[i].SupplierName)) + "|" +
			strings.ToLower(strings.TrimSpace(items[i].ProductName)) + "|" +
			strings.ToLower(strings.TrimSpace(items[i].Category)) + "|" +
			strings.ToLower(strings.TrimSpace(items[i].Subcategory))

		items[i].EquipmentID = deterministicSyntheticEquipmentID(signature)
	}
}

func deterministicSyntheticEquipmentID(signature string) uint {
	h := fnv.New32a()
	_, _ = h.Write([]byte(signature))
	// Keep synthetic IDs in a high range to avoid clashing with normal DB IDs.
	return uint(2_000_000_000 + (h.Sum32() % 1_000_000_000))
}

// GetDeviceTree fetches the WarehouseCore inventory tree used as a fallback
// source for product selection when RentalCore has no local inventory rows.
func (c *Client) GetDeviceTree() ([]DeviceTreeCategory, error) {
	url := fmt.Sprintf("%s/devices/tree", c.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create device tree request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch device tree: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device tree API returned status %d", resp.StatusCode)
	}

	var payload struct {
		TreeData []DeviceTreeCategory `json:"treeData"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode device tree: %w", err)
	}

	if payload.TreeData == nil {
		return []DeviceTreeCategory{}, nil
	}
	return payload.TreeData, nil
}

// GetActiveRentalEquipment fetches only active rental equipment
func (c *Client) GetActiveRentalEquipment() ([]RentalEquipmentItem, error) {
	items, err := c.GetRentalEquipment()
	if err != nil {
		return nil, err
	}

	// Filter active items
	active := make([]RentalEquipmentItem, 0, len(items))
	for _, item := range items {
		if item.IsActive {
			active = append(active, item)
		}
	}

	return active, nil
}

// GetRentalEquipmentBySupplier returns rental equipment grouped by supplier
func (c *Client) GetRentalEquipmentBySupplier() (map[string][]RentalEquipmentItem, error) {
	items, err := c.GetActiveRentalEquipment()
	if err != nil {
		return nil, err
	}

	// Group by supplier
	bySupplier := make(map[string][]RentalEquipmentItem)
	for _, item := range items {
		supplier := item.SupplierName
		if supplier == "" {
			supplier = "Unknown Supplier"
		}
		bySupplier[supplier] = append(bySupplier[supplier], item)
	}

	return bySupplier, nil
}

// RentalSubcategoryGroup groups rental items by subcategory within a category.
type RentalSubcategoryGroup struct {
	Subcategory string
	Items       []RentalEquipmentItem
}

// RentalCategoryGroup groups rental items by category.
type RentalCategoryGroup struct {
	Category      string
	Subcategories []RentalSubcategoryGroup
}

// GetRentalEquipmentByCategory returns rental equipment sorted by category → subcategory.
func (c *Client) GetRentalEquipmentByCategory() ([]RentalCategoryGroup, error) {
	items, err := c.GetActiveRentalEquipment()
	if err != nil {
		return nil, err
	}

	// category → subcategory → items
	type subKey struct{ cat, sub string }
	subMap := make(map[subKey][]RentalEquipmentItem)
	var catOrder []string
	catSeen := make(map[string]bool)
	subOrder := make(map[string][]string)
	subSeen := make(map[subKey]bool)

	for _, item := range items {
		rawCategory := strings.TrimSpace(item.Category)
		rawSubcategory := strings.TrimSpace(item.Subcategory)

		if rawCategory == "" && rawSubcategory != "" && strings.Contains(rawSubcategory, ">") {
			rawCategory = rawSubcategory
		}
		parts := strings.Split(rawCategory, ">")
		cleanParts := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				cleanParts = append(cleanParts, p)
			}
		}

		cat := "Other"
		if len(cleanParts) > 0 {
			cat = cleanParts[0]
		}

		sub := rawSubcategory
		if sub == "" && len(cleanParts) > 1 {
			sub = strings.Join(cleanParts[1:], " > ")
		} else if sub != "" && strings.Contains(sub, ">") {
			subParts := strings.Split(sub, ">")
			normSubParts := make([]string, 0, len(subParts))
			for _, p := range subParts {
				p = strings.TrimSpace(p)
				if p != "" {
					normSubParts = append(normSubParts, p)
				}
			}
			if cat == "Other" && len(normSubParts) > 0 {
				cat = normSubParts[0]
			}
			if len(normSubParts) > 1 {
				sub = strings.Join(normSubParts[1:], " > ")
			} else if len(normSubParts) == 1 {
				sub = normSubParts[0]
			}
		}
		if sub == "" {
			sub = strings.TrimSpace(item.SupplierName)
		}
		if sub == "" {
			sub = "General"
		}
		k := subKey{cat, sub}
		subMap[k] = append(subMap[k], item)
		if !catSeen[cat] {
			catSeen[cat] = true
			catOrder = append(catOrder, cat)
		}
		if !subSeen[k] {
			subSeen[k] = true
			subOrder[cat] = append(subOrder[cat], sub)
		}
	}

	sort.Strings(catOrder)
	for _, cat := range catOrder {
		sort.Strings(subOrder[cat])
		for _, sub := range subOrder[cat] {
			k := subKey{cat, sub}
			sort.SliceStable(subMap[k], func(i, j int) bool {
				return strings.ToLower(strings.TrimSpace(subMap[k][i].ProductName)) < strings.ToLower(strings.TrimSpace(subMap[k][j].ProductName))
			})
		}
	}

	result := make([]RentalCategoryGroup, 0, len(catOrder))
	for _, cat := range catOrder {
		group := RentalCategoryGroup{Category: cat}
		for _, sub := range subOrder[cat] {
			k := subKey{cat, sub}
			group.Subcategories = append(group.Subcategories, RentalSubcategoryGroup{
				Subcategory: sub,
				Items:       subMap[k],
			})
		}
		result = append(result, group)
	}
	return result, nil
}

// GetProduct fetches a single product by ID from WarehouseCore.
func (c *Client) GetProduct(id uint) (*Product, error) {
	url := fmt.Sprintf("%s/service/products/%d", c.baseURL, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create product request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch product %d: %w", id, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product %d not found", id)
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("WarehouseCore returned %d for product %d", resp.StatusCode, id)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching product %d", resp.StatusCode, id)
	}

	var p Product
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("decode product %d: %w", id, err)
	}
	return &p, nil
}

// ListProducts calls WarehouseCore product listing endpoint with optional search.
func (c *Client) ListProducts(search string) ([]Product, error) {
	url := fmt.Sprintf("%s/service/products", c.baseURL)
	if strings.TrimSpace(search) != "" {
		url = fmt.Sprintf("%s?search=%s", url, strings.ReplaceAll(strings.TrimSpace(search), " ", "%20"))
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create products request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch products: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("products API returned status %d", resp.StatusCode)
	}

	var items []Product
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode products list: %w", err)
	}
	return items, nil
}

// ListProductPackages calls WarehouseCore package listing endpoint.
func (c *Client) ListProductPackages(search string) ([]ProductPackage, error) {
	url := fmt.Sprintf("%s/service/product-packages", c.baseURL)
	if strings.TrimSpace(search) != "" {
		url = fmt.Sprintf("%s?search=%s", url, strings.ReplaceAll(strings.TrimSpace(search), " ", "%20"))
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create product packages request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch product packages: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product packages API returned status %d", resp.StatusCode)
	}

	var items []ProductPackage
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode product packages list: %w", err)
	}
	return items, nil
}

// GetProductPackage fetches a single product package with its items.
func (c *Client) GetProductPackage(packageID int) (*ProductPackageDetail, error) {
	url := fmt.Sprintf("%s/service/product-packages/%d", c.baseURL, packageID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create product package request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch product package %d: %w", packageID, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product package %d not found", packageID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product package detail API returned status %d", resp.StatusCode)
	}

	var item ProductPackageDetail
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("decode product package detail: %w", err)
	}
	return &item, nil
}

// GetCustomer fetches a single customer by ID from WarehouseCore.
func (c *Client) GetCustomer(id uint) (*Customer, error) {
	url := fmt.Sprintf("%s/admin/customers/%d", c.baseURL, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create customer request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch customer %d: %w", id, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("customer %d not found", id)
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("WarehouseCore returned %d for customer %d", resp.StatusCode, id)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching customer %d", resp.StatusCode, id)
	}

	var cust Customer
	if err := json.NewDecoder(resp.Body).Decode(&cust); err != nil {
		return nil, fmt.Errorf("decode customer %d: %w", id, err)
	}
	return &cust, nil
}

// ListCustomers calls WarehouseCore customer listing endpoint with optional search.
func (c *Client) ListCustomers(search string) ([]Customer, error) {
	url := fmt.Sprintf("%s/admin/customers", c.baseURL)
	if strings.TrimSpace(search) != "" {
		url = fmt.Sprintf("%s?search=%s", url, strings.ReplaceAll(strings.TrimSpace(search), " ", "%20"))
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create customers request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch customers: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("customers API returned status %d", resp.StatusCode)
	}

	var items []Customer
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode customers list: %w", err)
	}
	return items, nil
}

// GetUser fetches a single user by ID from the central auth API (RentalCore).
func (c *Client) GetUser(id uint) (*User, error) {
	url := fmt.Sprintf("%s/admin/users/%d", c.baseURL, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create user request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch user %d: %w", id, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user %d not found", id)
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("WarehouseCore returned %d for user %d", resp.StatusCode, id)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching user %d", resp.StatusCode, id)
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("decode user %d: %w", id, err)
	}
	return &u, nil
}

// ListUsers calls the central auth API user listing endpoint with optional search.
func (c *Client) ListUsers(search string) ([]User, error) {
	url := fmt.Sprintf("%s/admin/users", c.baseURL)
	if strings.TrimSpace(search) != "" {
		url = fmt.Sprintf("%s?search=%s", url, strings.ReplaceAll(strings.TrimSpace(search), " ", "%20"))
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create users request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.addAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch users: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users API returned status %d", resp.StatusCode)
	}

	var items []User
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode users list: %w", err)
	}
	return items, nil
}

// ClearCache clears the cached rental equipment data
func (c *Client) ClearCache() {
	c.mu.Lock()
	c.cache = nil
	c.cacheTime = time.Time{}
	c.mu.Unlock()
}
