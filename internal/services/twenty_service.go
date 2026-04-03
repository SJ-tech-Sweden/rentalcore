package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-barcode-webapp/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TwentyEnabledKey = "twenty.enabled"
	TwentyAPIURLKey  = "twenty.api_url"
	TwentyAPIKeyKey  = "twenty.api_key"
)

// TwentyService manages synchronisation between RentalCore and a Twenty CRM instance.
// Customer records are pushed to Twenty as Companies (company customers) or People
// (individual customers). Jobs are pushed as Opportunities.
// All remote calls are fire-and-forget (goroutine) so they never block the request path.
type TwentyService struct {
	db         *gorm.DB
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewTwentyService creates a new TwentyService.
func NewTwentyService(db *gorm.DB) *TwentyService {
	return &TwentyService{
		db: db,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// TwentyConfig holds the runtime configuration for the Twenty integration.
type TwentyConfig struct {
	Enabled bool
	APIURL  string
	APIKey  string
}

// GetConfig reads the current Twenty integration configuration from app_settings.
func (s *TwentyService) GetConfig() TwentyConfig {
	keys := []string{TwentyEnabledKey, TwentyAPIURLKey, TwentyAPIKeyKey}
	var settings []models.AppSetting
	s.db.Where("key IN ?", keys).Find(&settings)

	cfg := TwentyConfig{}
	for _, row := range settings {
		switch row.Key {
		case TwentyEnabledKey:
			cfg.Enabled = row.Value == "true"
		case TwentyAPIURLKey:
			cfg.APIURL = row.Value
		case TwentyAPIKeyKey:
			cfg.APIKey = row.Value
		}
	}
	return cfg
}

// SaveConfig persists the Twenty integration configuration to app_settings.
func (s *TwentyService) SaveConfig(cfg TwentyConfig) error {
	enabledVal := "false"
	if cfg.Enabled {
		enabledVal = "true"
	}
	rows := []models.AppSetting{
		{Key: TwentyEnabledKey, Value: enabledVal},
		{Key: TwentyAPIURLKey, Value: strings.TrimRight(cfg.APIURL, "/")},
		{Key: TwentyAPIKeyKey, Value: cfg.APIKey},
	}
	for _, row := range rows {
		if err := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"value": row.Value, "updated_at": time.Now()}),
		}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

// TestConnection verifies connectivity to the Twenty API and returns an error if it fails.
func (s *TwentyService) TestConnection() error {
	cfg := s.GetConfig()
	if cfg.APIURL == "" {
		return errors.New("twenty API URL is not configured")
	}
	if cfg.APIKey == "" {
		return errors.New("twenty API key is not configured")
	}

	// Use the metadata endpoint which returns the server's OpenAPI spec or a simple 200.
	// A lightweight query is used instead so we don't rely on a specific REST endpoint.
	payload, err := json.Marshal(gqlRequest{Query: "{ __typename }"})
	if err != nil {
		return fmt.Errorf("failed to encode test query: %w", err)
	}
	resp, err := s.doRequest(cfg, payload)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("authentication failed: invalid API key")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("twenty API returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// SyncCustomerAsync triggers an asynchronous best-effort sync of a customer to Twenty CRM.
func (s *TwentyService) SyncCustomerAsync(customer *models.Customer) {
	c := *customer // copy to avoid data race
	go func() {
		if err := s.syncCustomer(&c); err != nil {
			log.Printf("TwentyService: SyncCustomer for customer %d failed: %v", c.CustomerID, err)
		}
	}()
}

// SyncJobAsync triggers an asynchronous best-effort sync of a job to Twenty CRM.
func (s *TwentyService) SyncJobAsync(job *models.Job) {
	j := *job // copy to avoid data race
	go func() {
		if err := s.syncJob(&j); err != nil {
			log.Printf("TwentyService: SyncJob for job %d failed: %v", j.JobID, err)
		}
	}()
}

// syncCustomer pushes a customer to Twenty CRM.
// Company-type customers (or customers with a company name) are synced as Companies;
// individual customers are synced as People.
func (s *TwentyService) syncCustomer(customer *models.Customer) error {
	cfg := s.GetConfig()
	if !cfg.Enabled || cfg.APIURL == "" || cfg.APIKey == "" {
		return nil
	}

	isCompany := customer.CompanyName != nil && *customer.CompanyName != ""
	if isCompany {
		return s.upsertCompany(cfg, customer)
	}
	return s.upsertPerson(cfg, customer)
}

// syncJob pushes a job to Twenty CRM as an Opportunity.
func (s *TwentyService) syncJob(job *models.Job) error {
	cfg := s.GetConfig()
	if !cfg.Enabled || cfg.APIURL == "" || cfg.APIKey == "" {
		return nil
	}
	return s.upsertOpportunity(cfg, job)
}

// ---------- Twenty GraphQL helpers ----------

type gqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (s *TwentyService) doRequest(cfg TwentyConfig, body []byte) (*http.Response, error) {
	apiURL := cfg.APIURL + "/api"
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	return s.httpClient.Do(req)
}

func (s *TwentyService) execGQL(cfg TwentyConfig, query string, variables map[string]interface{}) (map[string]json.RawMessage, error) {
	payload, err := json.Marshal(gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return nil, err
	}

	resp, err := s.doRequest(cfg, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gqlResp gqlResponse
	if err := json.Unmarshal(raw, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode Twenty response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("Twenty API errors: %s", strings.Join(msgs, "; "))
	}
	return gqlResp.Data, nil
}

// storedID returns the Twenty object ID previously stored for a given RentalCore object, or "".
func (s *TwentyService) storedID(settingsKey string) string {
	var setting models.AppSetting
	if err := s.db.Where("key = ?", settingsKey).First(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}

// storeID persists a Twenty object ID mapping.
func (s *TwentyService) storeID(settingsKey, twentyID string) {
	row := models.AppSetting{Key: settingsKey, Value: twentyID}
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"value": twentyID, "updated_at": time.Now()}),
	}).Create(&row).Error; err != nil {
		log.Printf("TwentyService: failed to store ID mapping %q: %v", settingsKey, err)
	}
}

// ---------- Company sync ----------

func (s *TwentyService) upsertCompany(cfg TwentyConfig, c *models.Customer) error {
	mappingKey := fmt.Sprintf("twenty.company.%d", c.CustomerID)
	existingID := s.storedID(mappingKey)

	name := ""
	if c.CompanyName != nil {
		name = *c.CompanyName
	}
	if name == "" {
		name = c.GetDisplayName()
	}

	address := buildAddress(c)

	if existingID == "" {
		// Create
		const q = `
mutation CreateOneCompany($data: CompanyCreateInput!) {
  createCompany(data: $data) { id }
}`
		data := map[string]interface{}{
			"name":            name,
			"domainName":      map[string]interface{}{"primaryLinkUrl": "", "primaryLinkLabel": ""},
			"address":         address,
		}
		vars := map[string]interface{}{"data": data}
		resp, err := s.execGQL(cfg, q, vars)
		if err != nil {
			return err
		}
		id := extractID(resp, "createCompany")
		if id != "" {
			s.storeID(mappingKey, id)
		}
		return nil
	}

	// Update
	const q = `
mutation UpdateOneCompany($id: ID!, $data: CompanyUpdateInput!) {
  updateCompany(id: $id, data: $data) { id }
}`
	data := map[string]interface{}{
		"name":    name,
		"address": address,
	}
	vars := map[string]interface{}{"id": existingID, "data": data}
	_, err := s.execGQL(cfg, q, vars)
	return err
}

// ---------- Person sync ----------

func (s *TwentyService) upsertPerson(cfg TwentyConfig, c *models.Customer) error {
	mappingKey := fmt.Sprintf("twenty.person.%d", c.CustomerID)
	existingID := s.storedID(mappingKey)

	firstName := ""
	if c.FirstName != nil {
		firstName = *c.FirstName
	}
	lastName := ""
	if c.LastName != nil {
		lastName = *c.LastName
	}
	email := ""
	if c.Email != nil {
		email = *c.Email
	}
	phone := ""
	if c.PhoneNumber != nil {
		phone = *c.PhoneNumber
	}

	if existingID == "" {
		const q = `
mutation CreateOnePerson($data: PersonCreateInput!) {
  createPerson(data: $data) { id }
}`
		data := map[string]interface{}{
			"name":  map[string]interface{}{"firstName": firstName, "lastName": lastName},
			"emails": map[string]interface{}{"primaryEmail": email},
			"phones": map[string]interface{}{"primaryPhoneNumber": phone, "primaryPhoneCountryCode": ""},
		}
		vars := map[string]interface{}{"data": data}
		resp, err := s.execGQL(cfg, q, vars)
		if err != nil {
			return err
		}
		id := extractID(resp, "createPerson")
		if id != "" {
			s.storeID(mappingKey, id)
		}
		return nil
	}

	const q = `
mutation UpdateOnePerson($id: ID!, $data: PersonUpdateInput!) {
  updatePerson(id: $id, data: $data) { id }
}`
	data := map[string]interface{}{
		"name":   map[string]interface{}{"firstName": firstName, "lastName": lastName},
		"emails": map[string]interface{}{"primaryEmail": email},
		"phones": map[string]interface{}{"primaryPhoneNumber": phone, "primaryPhoneCountryCode": ""},
	}
	vars := map[string]interface{}{"id": existingID, "data": data}
	_, err := s.execGQL(cfg, q, vars)
	return err
}

// ---------- Opportunity sync ----------

func (s *TwentyService) upsertOpportunity(cfg TwentyConfig, job *models.Job) error {
	mappingKey := fmt.Sprintf("twenty.opportunity.%d", job.JobID)
	existingID := s.storedID(mappingKey)

	name := job.JobCode
	if name == "" {
		name = fmt.Sprintf("Job #%d", job.JobID)
	}

	// Build a description from the job description field if available.
	desc := ""
	if job.Description != nil {
		desc = *job.Description
	}

	// Determine stage from status
	stage := "NEW"
	if job.Status.Status != "" {
		stage = mapJobStatusToStage(job.Status.Status)
	}

	// Determine close date from end date
	closeDate := ""
	if job.EndDate != nil {
		closeDate = job.EndDate.Format("2006-01-02")
	}

	amount := job.Revenue
	if job.FinalRevenue != nil {
		amount = *job.FinalRevenue
	}

	if existingID == "" {
		const q = `
mutation CreateOneOpportunity($data: OpportunityCreateInput!) {
  createOpportunity(data: $data) { id }
}`
		data := map[string]interface{}{
			"name":        name,
			"stage":       stage,
			"amount":      map[string]interface{}{"amountMicros": int64(amount * 1_000_000), "currencyCode": "EUR"},
			"closeDate":   closeDate,
		}
		if desc != "" {
			data["pointOfContactNote"] = desc
		}
		vars := map[string]interface{}{"data": data}
		resp, err := s.execGQL(cfg, q, vars)
		if err != nil {
			return err
		}
		id := extractID(resp, "createOpportunity")
		if id != "" {
			s.storeID(mappingKey, id)
		}
		return nil
	}

	const q = `
mutation UpdateOneOpportunity($id: ID!, $data: OpportunityUpdateInput!) {
  updateOpportunity(id: $id, data: $data) { id }
}`
	data := map[string]interface{}{
		"name":      name,
		"stage":     stage,
		"amount":    map[string]interface{}{"amountMicros": int64(amount * 1_000_000), "currencyCode": "EUR"},
		"closeDate": closeDate,
	}
	if desc != "" {
		data["pointOfContactNote"] = desc
	}
	vars := map[string]interface{}{"id": existingID, "data": data}
	_, err := s.execGQL(cfg, q, vars)
	return err
}

// ---------- Helpers ----------

func buildAddress(c *models.Customer) map[string]interface{} {
	street := ""
	if c.Street != nil {
		street = *c.Street
	}
	if c.HouseNumber != nil && *c.HouseNumber != "" {
		street = strings.TrimSpace(street + " " + *c.HouseNumber)
	}
	city := ""
	if c.City != nil {
		city = *c.City
	}
	state := ""
	if c.FederalState != nil {
		state = *c.FederalState
	}
	zip := ""
	if c.ZIP != nil {
		zip = *c.ZIP
	}
	country := ""
	if c.Country != nil {
		country = *c.Country
	}
	return map[string]interface{}{
		"addressStreet1":    street,
		"addressCity":       city,
		"addressState":      state,
		"addressPostcode":   zip,
		"addressCountry":    country,
	}
}

func mapJobStatusToStage(status string) string {
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "new") || strings.Contains(lower, "open"):
		return "NEW"
	case strings.Contains(lower, "progress") || strings.Contains(lower, "active"):
		return "IN_PROGRESS"
	case strings.Contains(lower, "won") || strings.Contains(lower, "complet") || strings.Contains(lower, "done"):
		return "CLOSED_WON"
	case strings.Contains(lower, "lost") || strings.Contains(lower, "cancel"):
		return "CLOSED_LOST"
	default:
		return "NEW"
	}
}

// extractID pulls the "id" field out of a Twenty mutation response.
func extractID(data map[string]json.RawMessage, operationName string) string {
	raw, ok := data[operationName]
	if !ok {
		return ""
	}
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	return obj.ID
}
