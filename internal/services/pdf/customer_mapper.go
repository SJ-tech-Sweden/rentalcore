package pdf

import (
	"database/sql"
	"math"
	"strings"
	"time"

	"go-barcode-webapp/internal/models"
	"gorm.io/gorm"
)

const customerMatchThreshold = 60.0

// CustomerMapper handles customer mapping between PDF text and CRM customers
type CustomerMapper struct {
	DB *gorm.DB
}

// NewCustomerMapper creates a new customer mapper instance
func NewCustomerMapper(db *gorm.DB) *CustomerMapper {
	return &CustomerMapper{DB: db}
}

// FindBestMatch finds the best matching customer for given text
func (m *CustomerMapper) FindBestMatch(customerText string) (*models.PDFCustomerMapping, *models.Customer, float64, error) {
	if strings.TrimSpace(customerText) == "" {
		return nil, nil, 0, nil
	}

	var existing models.PDFCustomerMapping
	err := m.DB.Where("pdf_customer_text = ? AND is_active = ?", customerText, true).
		First(&existing).Error
	if err == nil {
		m.DB.Model(&existing).Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + 1"),
			"last_used_at": time.Now(),
		})

		var customer models.Customer
		if err := m.DB.First(&customer, existing.CustomerID).Error; err == nil {
			return &existing, &customer, 100.0, nil
		}
	}

	normalized := normalizeCustomerText(customerText)

	var customers []models.Customer
	if err := m.DB.Find(&customers).Error; err != nil {
		return nil, nil, 0, err
	}

	var bestMatch *models.Customer
	bestScore := 0.0

	for i := range customers {
		candidateName := customers[i].GetDisplayName()
		normalizedCandidate := normalizeCustomerText(candidateName)
		score := calculateSimilarity(normalized, normalizedCandidate)

		if normalizedCandidate != "" && (strings.Contains(normalizedCandidate, normalized) || strings.Contains(normalized, normalizedCandidate)) {
			score = math.Max(score, 90)
		}

		if customers[i].CompanyName != nil {
			companyNorm := normalizeCustomerText(*customers[i].CompanyName)
			if companyNorm == normalized {
				score = 100
			} else if companyNorm != "" && (strings.Contains(companyNorm, normalized) || strings.Contains(normalized, companyNorm)) {
				score = math.Max(score, 95)
			}
		}

		if fullName := normalizeCustomerText(buildCustomerFullName(&customers[i])); fullName != "" {
			fullScore := calculateSimilarity(normalized, fullName)
			if fullScore > score {
				score = fullScore
			}
			if strings.Contains(fullName, normalized) || strings.Contains(normalized, fullName) {
				score = math.Max(score, 85)
			}
		}

		if score > bestScore {
			bestScore = score
			bestMatch = &customers[i]
		}
	}

	if bestScore >= customerMatchThreshold && bestMatch != nil {
		return nil, bestMatch, bestScore, nil
	}

	return nil, nil, 0, nil
}

// SaveMapping saves a manual customer mapping
func (m *CustomerMapper) SaveMapping(pdfText string, customerID int, userID int64) error {
	if strings.TrimSpace(pdfText) == "" || customerID <= 0 {
		return nil
	}

	mapping := models.PDFCustomerMapping{
		PDFCustomerText: pdfText,
		NormalizedText:  sql.NullString{String: normalizeCustomerText(pdfText), Valid: true},
		CustomerID:      customerID,
		MappingType:     "manual",
		ConfidenceScore: sql.NullFloat64{Float64: 100.0, Valid: true},
		UsageCount:      1,
		LastUsedAt:      sql.NullTime{Time: time.Now(), Valid: true},
		CreatedBy:       sql.NullInt64{Int64: userID, Valid: true},
		IsActive:        true,
	}

	return m.DB.Create(&mapping).Error
}

func normalizeCustomerText(text string) string {
	text = strings.ToLower(text)
	text = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return -1
	}, text)
	fields := strings.Fields(text)
	filtered := make([]string, 0, len(fields))
	for _, field := range fields {
		if shouldSkipCustomerToken(field) {
			continue
		}
		filtered = append(filtered, field)
	}
	text = strings.Join(filtered, " ")
	return strings.TrimSpace(text)
}

func shouldSkipCustomerToken(token string) bool {
	switch token {
	case "gmbh", "mbh", "ug", "kg", "ag", "ltd", "inc", "co", "und", "&", "eventtechnik", "events", "event", "verleih":
		return true
	}
	return false
}

func buildCustomerFullName(customer *models.Customer) string {
	if customer == nil {
		return ""
	}
	parts := make([]string, 0, 2)
	if customer.FirstName != nil && strings.TrimSpace(*customer.FirstName) != "" {
		parts = append(parts, *customer.FirstName)
	}
	if customer.LastName != nil && strings.TrimSpace(*customer.LastName) != "" {
		parts = append(parts, *customer.LastName)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}
