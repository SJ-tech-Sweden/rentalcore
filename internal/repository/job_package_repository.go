package repository

import (
	"database/sql"
	"fmt"
	"go-barcode-webapp/internal/models"
	"log"
	"time"

	"gorm.io/gorm"
)

type JobPackageRepository struct {
	db *Database
}

func NewJobPackageRepository(db *Database) *JobPackageRepository {
	return &JobPackageRepository{db: db}
}

// AssignPackageToJob assigns package devices to a job
// v5.0: Only records the package on the job (no package devices, expansion handled elsewhere)
func (r *JobPackageRepository) AssignPackageToJob(jobID int, packageID int, quantity uint, customPrice *float64, userID uint) (*models.JobPackage, error) {
	log.Printf("=== AssignPackageToJob v5.0 START: jobID=%d, packageID=%d, qty=%d ===", jobID, packageID, quantity)

	defaultPrice := sql.NullFloat64{Valid: false}
	foundPackage := false

	if r.db.Migrator().HasTable(&models.ProductPackage{}) {
		var pkg models.ProductPackage
		if err := r.db.Where("package_id = ?", packageID).First(&pkg).Error; err == nil {
			foundPackage = true
			if pkg.Price.Valid {
				defaultPrice = sql.NullFloat64{Float64: pkg.Price.Float64, Valid: true}
			}
		}
	}

	if !foundPackage && r.db.Migrator().HasTable(&models.EquipmentPackage{}) {
		// Fallback for installations still using equipment_packages.
		var legacyPkg models.EquipmentPackage
		if fallbackErr := r.db.Where("packageID = ?", packageID).First(&legacyPkg).Error; fallbackErr == nil {
			foundPackage = true
			if legacyPkg.PackagePrice != nil {
				defaultPrice = sql.NullFloat64{Float64: *legacyPkg.PackagePrice, Valid: true}
			}
		}
	}

	if !foundPackage {
		log.Printf("⚠️ package %d not found in local package tables; recording assignment without local package metadata", packageID)
	}

	// Build price value
	var priceValue sql.NullFloat64
	if customPrice != nil {
		priceValue = sql.NullFloat64{Float64: *customPrice, Valid: true}
	} else if defaultPrice.Valid {
		priceValue = defaultPrice
	}

	jobPackage := &models.JobPackage{
		JobID:       jobID,
		PackageID:   packageID,
		Quantity:    quantity,
		CustomPrice: priceValue,
		AddedAt:     time.Now(),
		AddedBy:     &userID,
	}

	var existing models.JobPackage
	err := r.db.Where("job_id = ? AND package_id = ?", jobID, packageID).First(&existing).Error
	if err == nil {
		existing.Quantity = quantity
		existing.CustomPrice = priceValue
		existing.AddedAt = time.Now()
		existing.AddedBy = &userID
		if saveErr := r.db.Save(&existing).Error; saveErr != nil {
			return nil, fmt.Errorf("failed to update job_package: %w", saveErr)
		}
		jobPackage = &existing
	} else if err == gorm.ErrRecordNotFound {
		if createErr := r.db.Create(jobPackage).Error; createErr != nil {
			return nil, fmt.Errorf("failed to create job_package: %w", createErr)
		}
	} else {
		return nil, fmt.Errorf("failed to check existing job_package: %w", err)
	}

	log.Printf("=== AssignPackageToJob v5.0 RECORDED package %d on job %d (qty=%d, price=%v) ===", packageID, jobID, quantity, priceValue)
	return jobPackage, nil
}

// checkPackageItemAvailability verifies all products in package are available
func (r *JobPackageRepository) checkPackageItemAvailability(tx *Database, packageItems []models.ProductPackageItem, quantity uint, startDate, endDate sql.NullTime, excludeJobID int) []string {
	var issues []string

	for _, pkgItem := range packageItems {
		totalNeeded := uint(pkgItem.Quantity) * quantity
		available, err := r.countAvailableDevicesByProduct(tx, pkgItem.ProductID, startDate, endDate, excludeJobID)

		// Load product name for better error message
		var product models.Product
		productName := fmt.Sprintf("Product ID %d", pkgItem.ProductID)
		if err := tx.Where("productID = ?", pkgItem.ProductID).First(&product).Error; err == nil {
			productName = product.Name
		}

		if err != nil {
			issues = append(issues, fmt.Sprintf("Error checking %s: %v", productName, err))
			continue
		}

		if available < totalNeeded {
			issues = append(issues, fmt.Sprintf("%s: need %d, only %d available", productName, totalNeeded, available))
		}
	}

	return issues
}

// countAvailableDevicesByProduct counts how many devices of a product are available
func (r *JobPackageRepository) countAvailableDevicesByProduct(tx *Database, productID int, startDate, endDate sql.NullTime, excludeJobID int) (uint, error) {
	// Count total devices of this product type
	var totalCount int64
	if err := tx.Model(&models.Device{}).Where("productID = ?", productID).Count(&totalCount).Error; err != nil {
		return 0, err
	}

	// If no date range, return total
	if !startDate.Valid || !endDate.Valid {
		return uint(totalCount), nil
	}

	// Count devices already reserved in overlapping jobs
	var reservedCount int64
	query := `
		SELECT COUNT(DISTINCT d.deviceid)
		FROM devices d
		WHERE d.productid = ?
		AND (
			EXISTS (
				SELECT 1 FROM job_devices jd
				JOIN jobs j ON jd.jobid COLLATE utf8mb4_unicode_ci = j.jobid COLLATE utf8mb4_unicode_ci
				WHERE jd.deviceid COLLATE utf8mb4_unicode_ci = d.deviceid COLLATE utf8mb4_unicode_ci
				AND j.jobid != ?
				AND j.startdate IS NOT NULL
				AND j.enddate IS NOT NULL
				AND j.startdate <= ?
				AND j.enddate >= ?
			)
			OR EXISTS (
				SELECT 1 FROM job_package_reservations jpr
				JOIN job_packages jp ON jpr.job_package_id = jp.job_package_id
				JOIN jobs j ON jp.job_id = j.jobid
				WHERE jpr.device_id COLLATE utf8mb4_unicode_ci = d.deviceid COLLATE utf8mb4_unicode_ci
				AND jpr.reservation_status = 'reserved'
				AND j.jobid != ?
				AND j.startdate IS NOT NULL
				AND j.enddate IS NOT NULL
				AND j.startdate <= ?
				AND j.enddate >= ?
			)
		)
	`
	if err := tx.Raw(query, productID, excludeJobID, endDate.Time, startDate.Time, excludeJobID, endDate.Time, startDate.Time).Scan(&reservedCount).Error; err != nil {
		return 0, err
	}

	available := totalCount - reservedCount
	if available < 0 {
		available = 0
	}

	return uint(available), nil
}

// findAvailableDevicesByProduct finds specific device instances by product that are available
func (r *JobPackageRepository) findAvailableDevicesByProduct(tx *Database, productID int, quantity uint, startDate, endDate sql.NullTime, excludeJobID int) ([]string, error) {
	var devices []string

	query := `
		SELECT d.deviceid
		FROM devices d
		WHERE d.productid = ?
		AND NOT EXISTS (
			SELECT 1 FROM job_devices jd
			JOIN jobs j ON jd.jobid COLLATE utf8mb4_unicode_ci = j.jobid COLLATE utf8mb4_unicode_ci
			WHERE jd.deviceid COLLATE utf8mb4_unicode_ci = d.deviceid COLLATE utf8mb4_unicode_ci
			AND j.jobid != ?
			AND j.startdate IS NOT NULL
			AND j.enddate IS NOT NULL
			AND j.startdate <= ?
			AND j.enddate >= ?
		)
		AND NOT EXISTS (
			SELECT 1 FROM job_package_reservations jpr
			JOIN job_packages jp ON jpr.job_package_id = jp.job_package_id
			JOIN jobs j ON jp.job_id = j.jobid
			WHERE jpr.device_id COLLATE utf8mb4_unicode_ci = d.deviceid COLLATE utf8mb4_unicode_ci
			AND jpr.reservation_status = 'reserved'
			AND j.jobid != ?
			AND j.startdate IS NOT NULL
			AND j.enddate IS NOT NULL
			AND j.startdate <= ?
			AND j.enddate >= ?
		)
		LIMIT ?
	`

	if err := tx.Raw(query, productID, excludeJobID, endDate.Time, startDate.Time, excludeJobID, endDate.Time, startDate.Time, quantity).Scan(&devices).Error; err != nil {
		return nil, err
	}

	if len(devices) < int(quantity) {
		return nil, fmt.Errorf("insufficient available devices: need %d, found %d", quantity, len(devices))
	}

	return devices, nil
}

// GetJobPackageByID retrieves a job package by ID with associations
func (r *JobPackageRepository) GetJobPackageByID(id uint) (*models.JobPackage, error) {
	var jobPackage models.JobPackage
	query := r.db.
		Preload("AddedByUser").
		Where("job_package_id = ?", id)

	if r.db.Migrator().HasTable(&models.JobPackageReservation{}) {
		query = query.Preload("Reservations").Preload("Reservations.Device")
	}

	if r.db.Migrator().HasTable(&models.ProductPackage{}) {
		query = query.Preload("Package")
	}

	err := query.First(&jobPackage).Error

	if err != nil {
		return nil, err
	}

	return &jobPackage, nil
}

// GetJobPackagesByJobID retrieves all packages for a job
func (r *JobPackageRepository) GetJobPackagesByJobID(jobID int) ([]models.JobPackage, error) {
	var packages []models.JobPackage
	query := r.db.
		Preload("AddedByUser").
		Where("job_id = ?", jobID).
		Order("added_at DESC")

	if r.db.Migrator().HasTable(&models.JobPackageReservation{}) {
		query = query.Preload("Reservations").Preload("Reservations.Device")
	}

	if r.db.Migrator().HasTable(&models.ProductPackage{}) {
		query = query.Preload("Package")
	}

	err := query.Find(&packages).Error

	return packages, err
}

// UpdateJobPackageQuantity updates the quantity of a job package
func (r *JobPackageRepository) UpdateJobPackageQuantity(jobPackageID uint, newQuantity uint) error {
	// This would require re-calculating reservations
	// For now, we'll implement a simple version
	return r.db.Model(&models.JobPackage{}).
		Where("job_package_id = ?", jobPackageID).
		Update("quantity", newQuantity).Error
}

// UpdateJobPackagePrice updates the custom price
func (r *JobPackageRepository) UpdateJobPackagePrice(jobPackageID uint, newPrice *float64) error {
	var priceValue sql.NullFloat64
	if newPrice != nil {
		priceValue = sql.NullFloat64{Float64: *newPrice, Valid: true}
	} else {
		priceValue = sql.NullFloat64{Valid: false}
	}

	return r.db.Model(&models.JobPackage{}).
		Where("job_package_id = ?", jobPackageID).
		Update("custom_price", priceValue).Error
}

// RemoveJobPackage removes a package from a job and releases reservations
func (r *JobPackageRepository) RemoveJobPackage(jobPackageID uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update reservation status to released when reservation table exists.
	if tx.Migrator().HasTable(&models.JobPackageReservation{}) {
		if err := tx.Model(&models.JobPackageReservation{}).
			Where("job_package_id = ?", jobPackageID).
			Updates(map[string]interface{}{
				"reservation_status": "released",
				"released_at":        time.Now(),
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete the job package (cascades to reservations via DB constraints)
	if err := tx.Delete(&models.JobPackage{}, jobPackageID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetPackageDeviceReservations retrieves all device reservations for a package
func (r *JobPackageRepository) GetPackageDeviceReservations(jobPackageID uint) ([]models.JobPackageReservation, error) {
	if !r.db.Migrator().HasTable(&models.JobPackageReservation{}) {
		return []models.JobPackageReservation{}, nil
	}

	var reservations []models.JobPackageReservation
	err := r.db.
		Preload("Device").
		Preload("Device.Product").
		Where("job_package_id = ?", jobPackageID).
		Order("reserved_at").
		Find(&reservations).Error

	return reservations, err
}

// ReleasePackageReservations releases all device reservations for a package
func (r *JobPackageRepository) ReleasePackageReservations(jobPackageID uint) error {
	if !r.db.Migrator().HasTable(&models.JobPackageReservation{}) {
		return nil
	}

	return r.db.Model(&models.JobPackageReservation{}).
		Where("job_package_id = ? AND reservation_status = 'reserved'", jobPackageID).
		Updates(map[string]interface{}{
			"reservation_status": "released",
			"released_at":        time.Now(),
		}).Error
}

// GetJobPackagesWithDetails retrieves packages with computed details for display
func (r *JobPackageRepository) GetJobPackagesWithDetails(jobID int) ([]models.JobPackageWithDetails, error) {
	hasProductPackages := r.db.Migrator().HasTable(&models.ProductPackage{})
	hasProductPackageItems := r.db.Migrator().HasTable(&models.ProductPackageItem{})
	hasReservations := r.db.Migrator().HasTable(&models.JobPackageReservation{})

	var packages []models.JobPackage
	query := r.db.
		Where("job_id = ?", jobID).
		Order("added_at DESC")
	if hasReservations {
		query = query.Preload("Reservations")
	}
	if hasProductPackages {
		query = query.Preload("Package")
	}

	err := query.Find(&packages).Error

	if err != nil {
		return nil, err
	}

	result := make([]models.JobPackageWithDetails, len(packages))
	for i, pkg := range packages {
		details := models.JobPackageWithDetails{
			JobPackage: pkg,
		}

		if pkg.Package != nil {
			details.PackageName = pkg.Package.Name
			if pkg.Package.Description.Valid {
				details.PackageDescription = pkg.Package.Description.String
			}
			if pkg.Package.Price.Valid {
				details.PackagePrice = pkg.Package.Price.Float64
			}

			// Count items in the package from product_package_items
			if hasProductPackageItems {
				var itemCount int64
				r.db.Model(&models.ProductPackageItem{}).Where("package_id = ?", pkg.Package.PackageID).Count(&itemCount)
				details.DeviceCount = int(itemCount)
			}
		} else {
			details.PackageName = fmt.Sprintf("Package #%d", pkg.PackageID)
		}

		// Calculate effective price
		if pkg.CustomPrice.Valid {
			details.EffectivePrice = pkg.CustomPrice.Float64
		} else {
			details.EffectivePrice = details.PackagePrice
		}

		details.TotalPrice = details.EffectivePrice * float64(pkg.Quantity)
		details.ReservedDevices = len(pkg.Reservations)

		// Determine availability status
		if details.ReservedDevices >= details.DeviceCount*int(pkg.Quantity) {
			details.AvailabilityStatus = "fully_reserved"
		} else if details.ReservedDevices > 0 {
			details.AvailabilityStatus = "partially_reserved"
		} else {
			details.AvailabilityStatus = "not_reserved"
		}

		result[i] = details
	}

	return result, nil
}
