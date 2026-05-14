package repository

import (
	"errors"
	"fmt"
	"log"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/services/warehousecore"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RentalEquipmentRepository struct {
	db       *Database
	whClient *warehousecore.Client
}

func NewRentalEquipmentRepository(db *Database) *RentalEquipmentRepository {
	return &RentalEquipmentRepository{db: db}
}

// WithWarehouseCoreClient attaches a WarehouseCore client and enables API-based reads.
func (r *RentalEquipmentRepository) WithWarehouseCoreClient(client *warehousecore.Client) *RentalEquipmentRepository {
	r.whClient = client
	return r
}

// GetAllRentalEquipment returns all rental equipment items
func (r *RentalEquipmentRepository) GetAllRentalEquipment(rentalEquipment *[]models.RentalEquipment) error {
	// Prefer WarehouseCore API when configured
	if r.whClient != nil {
		items, err := r.whClient.GetActiveRentalEquipment()
		if err == nil {
			out := make([]models.RentalEquipment, 0, len(items))
			for _, it := range items {
				out = append(out, models.RentalEquipment{
					EquipmentID:  it.EquipmentID,
					ProductName:  it.ProductName,
					SupplierName: it.SupplierName,
					RentalPrice:  it.RentalPrice,
					Category:     it.Category,
					Description:  it.Description,
					IsActive:     it.IsActive,
				})
			}
			*rentalEquipment = out
			return nil
		}
		// Fall back to DB on API error
	}

	return r.db.Find(rentalEquipment).Error
}

// GetRentalEquipmentByID returns a specific rental equipment item by ID
func (r *RentalEquipmentRepository) GetRentalEquipmentByID(equipmentID uint, rentalEquipment *models.RentalEquipment) error {
	if r.whClient != nil {
		// Try WarehouseCore first
		items, err := r.whClient.GetActiveRentalEquipment()
		if err == nil {
			for _, it := range items {
				if it.EquipmentID == equipmentID {
					*rentalEquipment = models.RentalEquipment{
						EquipmentID:  it.EquipmentID,
						ProductName:  it.ProductName,
						SupplierName: it.SupplierName,
						RentalPrice:  it.RentalPrice,
						Category:     it.Category,
						Description:  it.Description,
						IsActive:     it.IsActive,
					}
					return nil
				}
			}
		}
		// Fall through to DB
	}

	return r.db.First(rentalEquipment, equipmentID).Error
}

// CreateRentalEquipment creates a new rental equipment item
func (r *RentalEquipmentRepository) CreateRentalEquipment(rentalEquipment *models.RentalEquipment) error {
	// Creation of rental equipment is now managed in WarehouseCore
	return fmt.Errorf("rental equipment management moved to WarehouseCore; create via WarehouseCore admin API")
}

// UpdateRentalEquipment updates an existing rental equipment item
func (r *RentalEquipmentRepository) UpdateRentalEquipment(rentalEquipment *models.RentalEquipment) error {
	return fmt.Errorf("rental equipment management moved to WarehouseCore; update via WarehouseCore admin API")
}

// DeleteRentalEquipment deletes a rental equipment item
func (r *RentalEquipmentRepository) DeleteRentalEquipment(equipmentID uint) error {
	return fmt.Errorf("rental equipment management moved to WarehouseCore; delete via WarehouseCore admin API")
}

// UpsertEquipmentFromWC ensures a WarehouseCore-managed rental equipment record
// exists in the local rental_equipment table so that job_rental_equipment FK
// constraints are satisfied. It creates a minimal record or updates the price/name
// if already present.
func (r *RentalEquipmentRepository) UpsertEquipmentFromWC(equipmentID uint, productName, supplierName string, rentalPrice float64, category string) error {
	if !r.db.Migrator().HasTable("rental_equipment") {
		log.Printf("⚠️ Warning: rental_equipment table does not exist; skipping upsert for equipment_id %d", equipmentID)
		return nil
	}

	if supplierName == "" {
		supplierName = "External"
	}
	equipment := models.RentalEquipment{
		EquipmentID:  equipmentID,
		ProductName:  productName,
		SupplierName: supplierName,
		RentalPrice:  rentalPrice,
		Category:     category,
		IsActive:     true,
	}
	// Use GORM clauses for a proper upsert (INSERT … ON CONFLICT DO UPDATE)
	return r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "equipment_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"product_name", "supplier_name", "rental_price", "category", "updated_at"}),
		}).
		Create(&equipment).Error
}

// AddRentalToJob adds rental equipment to a job
// Gracefully handles missing table (migration not run yet)
func (r *RentalEquipmentRepository) AddRentalToJob(jobRental *models.JobRentalEquipment) error {
	if !r.db.Migrator().HasTable("job_rental_equipment") || !r.db.Migrator().HasTable("rental_equipment") {
		log.Printf("⚠️ Warning: rental job link tables missing; skipping rental equipment addition for job_id %d", jobRental.JobID)
		return nil
	}

	// Get rental equipment to calculate total cost
	var equipment models.RentalEquipment
	if err := r.db.First(&equipment, jobRental.EquipmentID).Error; err != nil {
		return fmt.Errorf("rental equipment not found: %v", err)
	}

	// Calculate total cost
	jobRental.TotalCost = equipment.RentalPrice * float64(jobRental.Quantity) * float64(jobRental.DaysUsed)

	// Check if already exists, then update or create
	var existingRental models.JobRentalEquipment
	err := r.db.Where("job_id = ? AND equipment_id = ?", jobRental.JobID, jobRental.EquipmentID).First(&existingRental).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		createErr := r.db.Create(jobRental).Error
		if createErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(createErr, &pgErr) && pgErr.Code == "42P01" {
				log.Printf("⚠️ Warning: job_rental_equipment table does not exist (42P01); skipping rental equipment addition")
				return nil
			}
		}
		return createErr
	} else if err != nil {
		// Check if this is a missing table error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			log.Printf("⚠️ Warning: job_rental_equipment table does not exist (42P01); skipping rental equipment addition")
			return nil
		}
		return err
	} else {
		// Update existing
		existingRental.Quantity = jobRental.Quantity
		existingRental.DaysUsed = jobRental.DaysUsed
		existingRental.TotalCost = jobRental.TotalCost
		existingRental.Notes = jobRental.Notes
		updateErr := r.db.Save(&existingRental).Error
		if updateErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(updateErr, &pgErr) && pgErr.Code == "42P01" {
				log.Printf("⚠️ Warning: job_rental_equipment table does not exist (42P01); skipping rental equipment update")
				return nil
			}
		}
		return updateErr
	}
}

// CreateRentalEquipmentFromManualEntry creates rental equipment and adds it to job in one transaction
func (r *RentalEquipmentRepository) CreateRentalEquipmentFromManualEntry(request *models.ManualRentalEntryRequest, createdBy *uint) (*models.RentalEquipment, *models.JobRentalEquipment, error) {
	var rentalEquipment *models.RentalEquipment
	var jobRental *models.JobRentalEquipment

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Create rental equipment
		rentalEquipment = &models.RentalEquipment{
			ProductName:  request.ProductName,
			SupplierName: request.SupplierName,
			RentalPrice:  request.RentalPrice,
			Category:     request.Category,
			Description:  request.Description,
			Notes:        request.Notes,
			IsActive:     true,
			CreatedBy:    createdBy,
		}

		if err := tx.Create(rentalEquipment).Error; err != nil {
			return fmt.Errorf("failed to create rental equipment: %v", err)
		}

		// Add to job
		totalCost := request.RentalPrice * float64(request.Quantity) * float64(request.DaysUsed)

		jobRental = &models.JobRentalEquipment{
			JobID:       request.JobID,
			EquipmentID: rentalEquipment.EquipmentID,
			Quantity:    request.Quantity,
			DaysUsed:    request.DaysUsed,
			TotalCost:   totalCost,
			Notes:       request.Notes,
		}

		if err := tx.Create(jobRental).Error; err != nil {
			return fmt.Errorf("failed to add rental to job: %v", err)
		}

		return nil
	})

	return rentalEquipment, jobRental, err
}

// GetJobRentalEquipment returns all rental equipment for a specific job
// Returns an empty slice gracefully if the table doesn't exist (migration not run)
func (r *RentalEquipmentRepository) GetJobRentalEquipment(jobID uint, jobRentals *[]models.JobRentalEquipment) error {
	if !r.db.Migrator().HasTable("job_rental_equipment") {
		*jobRentals = []models.JobRentalEquipment{}
		return nil
	}

	err := r.db.Preload("RentalEquipment").Where("job_id = ?", jobID).Find(jobRentals).Error

	// Handle missing table (42P01 error) gracefully
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			// Table doesn't exist (migration not run yet)
			log.Printf("⚠️ Warning: job_rental_equipment table does not exist (42P01); migration may not have run. Returning empty list.")
			*jobRentals = []models.JobRentalEquipment{}
			return nil
		}
		// Return any other errors as-is
		return err
	}
	return nil
}

// RemoveRentalFromJob removes rental equipment from a job
// Gracefully handles missing table (migration not run yet)
func (r *RentalEquipmentRepository) RemoveRentalFromJob(jobID, equipmentID uint) error {
	err := r.db.Where("job_id = ? AND equipment_id = ?", jobID, equipmentID).Delete(&models.JobRentalEquipment{}).Error

	// Handle missing table gracefully
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			log.Printf("⚠️ Warning: job_rental_equipment table does not exist (42P01); skipping removal")
			return nil
		}
		return err
	}
	return nil
}

// GetRentalEquipmentAnalytics returns analytics data for rental equipment
func (r *RentalEquipmentRepository) GetRentalEquipmentAnalytics() (*models.RentalEquipmentAnalytics, error) {
	analytics := &models.RentalEquipmentAnalytics{}

	// Basic counts
	var totalCount, activeCount int64
	r.db.Model(&models.RentalEquipment{}).Count(&totalCount)
	r.db.Model(&models.RentalEquipment{}).Where("is_active = ?", true).Count(&activeCount)
	analytics.TotalEquipmentItems = int(totalCount)
	analytics.ActiveEquipmentItems = int(activeCount)

	// Count distinct suppliers
	var suppliers []string
	r.db.Model(&models.RentalEquipment{}).Distinct("supplier_name").Pluck("supplier_name", &suppliers)
	analytics.TotalSuppliersCount = len(suppliers)

	// Total rental revenue
	r.db.Model(&models.JobRentalEquipment{}).Select("COALESCE(SUM(total_cost), 0)").Scan(&analytics.TotalRentalRevenue)

	// Basic category breakdown (simplified for now)
	var categories []models.RentalCategoryBreakdown

	// Get categories with equipment count
	type CategorySummary struct {
		Category       string
		EquipmentCount int64
		TotalRevenue   float64
	}

	var categorySummaries []CategorySummary
	r.db.Model(&models.RentalEquipment{}).
		Select("COALESCE(category, 'Uncategorized') as category, COUNT(*) as equipment_count").
		Group("category").
		Find(&categorySummaries)

	for _, summary := range categorySummaries {
		var avgRevenue float64
		if summary.EquipmentCount > 0 {
			avgRevenue = summary.TotalRevenue / float64(summary.EquipmentCount)
		}

		categories = append(categories, models.RentalCategoryBreakdown{
			Category:               summary.Category,
			EquipmentCount:         int(summary.EquipmentCount),
			TotalRevenue:           summary.TotalRevenue,
			UsageCount:             0, // Simplified for now
			AvgRevenuePerEquipment: avgRevenue,
		})
	}

	// For simplicity, using simplified data for now
	analytics.MostUsedEquipment = []models.MostUsedRentalEquipment{}
	analytics.TopSuppliers = []models.TopRentalSupplier{}
	analytics.CategoryBreakdown = categories
	analytics.MonthlyRentalRevenue = []models.MonthlyRentalRevenue{}

	return analytics, nil
}
