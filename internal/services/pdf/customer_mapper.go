package pdf

import (
	"database/sql"
	"strings"
	"time"

	"go-barcode-webapp/internal/models"
	"gorm.io/gorm"
)

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
		name := customers[i].GetDisplayName()
		score := calculateSimilarity(normalized, normalizeCustomerText(name))
		if score > bestScore {
			bestScore = score
			bestMatch = &customers[i]
		}
	}

	if bestScore >= 70.0 && bestMatch != nil {
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
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}
