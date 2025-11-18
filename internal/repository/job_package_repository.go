package repository

import (
	"database/sql"
	"fmt"
	"go-barcode-webapp/internal/models"
	"log"
	"strings"
	"time"
)

type JobPackageRepository struct {
	db *Database
}

func NewJobPackageRepository(db *Database) *JobPackageRepository {
	return &JobPackageRepository{db: db}
}

// AssignPackageToJob assigns package devices to a job
// v4.0 SIMPLIFIED: Works with manually created package devices (e.g., PKG_SOUNDM_001)
// This is a compatibility layer for OCR - packages are now treated like regular devices
func (r *JobPackageRepository) AssignPackageToJob(jobID int, packageID int, quantity uint, customPrice *float64, userID uint) (*models.JobPackage, error) {
	log.Printf("=== AssignPackageToJob v4.0 START: jobID=%d, packageID=%d, qty=%d ===", jobID, packageID, quantity)

	// Verify package exists
	var pkg models.ProductPackage
	if err := r.db.Where("package_id = ?", packageID).First(&pkg).Error; err != nil {
		return nil, fmt.Errorf("package %d not found: %w", packageID, err)
	}
	log.Printf("[v4.0] Package found: %s", pkg.Name)

	// Get job for date range
	var job models.Job
	if err := r.db.First(&job, jobID).Error; err != nil {
		return nil, fmt.Errorf("job %d not found: %w", jobID, err)
	}

	// Find available package devices (devices whose productID = packageID)
	// Package devices are regular devices, but their product is a package
	var availablePackageDevices []models.Device
	query := r.db.Model(&models.Device{}).Where("productID = ?", packageID)

	// Exclude devices assigned to other jobs with overlapping dates
	if job.StartDate != nil && job.EndDate != nil {
		query = query.Where(`deviceID NOT IN (
			SELECT jd.deviceID FROM jobdevices jd
			JOIN jobs j ON jd.jobID = j.jobID
			WHERE j.startDate <= ? AND j.endDate >= ?
			AND j.statusID IN (SELECT statusID FROM status WHERE status IN ('open', 'in_progress'))
		)`, job.EndDate, job.StartDate)
	}

	query = query.Order("deviceID ASC").Limit(int(quantity))

	if err := query.Find(&availablePackageDevices).Error; err != nil {
		return nil, fmt.Errorf("failed to find package devices: %w", err)
	}

	if len(availablePackageDevices) < int(quantity) {
		return nil, fmt.Errorf("not enough available package devices: need %d, found %d", quantity, len(availablePackageDevices))
	}

	log.Printf("[v4.0] Found %d available package devices", len(availablePackageDevices))

	// Calculate price per package device
	var pricePerDevice *float64
	if customPrice != nil && *customPrice > 0 && quantity > 0 {
		price := *customPrice / float64(quantity)
		pricePerDevice = &price
	}

	// Start transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Add each package device to jobdevices
	// Then manually trigger package assignment logic (adding real devices with discounts)
	for i, device := range availablePackageDevices {
		price := float64(0)
		if pricePerDevice != nil {
			price = *pricePerDevice
		}

		// Load package items
		var packageItems []models.ProductPackageItem
		if err := tx.Where("package_id = ?", packageID).Find(&packageItems).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to load package items: %w", err)
		}

		// Calculate total regular price
		var regularTotal float64
		for _, item := range packageItems {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				continue
			}
			if product.ItemCostPerDay != nil {
				regularTotal += *product.ItemCostPerDay * float64(item.Quantity)
			}
		}

		// Calculate discount percentage
		packagePrice := price
		if packagePrice == 0 && pkg.Price.Valid {
			packagePrice = pkg.Price.Float64
		}
		var discountPercent float64
		if regularTotal > 0 {
			discountPercent = (regularTotal - packagePrice) / regularTotal
		}

		// Add package device itself
		packageJobDevice := models.JobDevice{
			JobID:       jobID,
			DeviceID:    device.DeviceID,
			CustomPrice: &price,
		}
		if err := tx.Create(&packageJobDevice).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to assign package device %s: %w", device.DeviceID, err)
		}

		log.Printf("[v4.0] ✓ Assigned package device %d/%d: %s (price: %.2f, discount: %.1f%%)",
			i+1, quantity, device.DeviceID, price, discountPercent*100)

		// Add real devices from the package with discounted prices
		for _, item := range packageItems {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				continue
			}

			// Find available devices for this product
			var availableDevices []models.Device
			deviceQuery := tx.Model(&models.Device{}).Where("productID = ?", item.ProductID)
			if job.StartDate != nil && job.EndDate != nil {
				deviceQuery = deviceQuery.Where(`deviceID NOT IN (
					SELECT jd.deviceID FROM jobdevices jd
					JOIN jobs j ON jd.jobID = j.jobID
					WHERE j.startDate <= ? AND j.endDate >= ?
					AND j.statusID IN (SELECT statusID FROM status WHERE status IN ('open', 'in_progress'))
				)`, job.EndDate, job.StartDate)
			}
			deviceQuery = deviceQuery.Order("serialnumber ASC").Limit(item.Quantity)

			if err := deviceQuery.Find(&availableDevices).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to find devices for product %d: %w", item.ProductID, err)
			}

			if len(availableDevices) < item.Quantity {
				tx.Rollback()
				return nil, fmt.Errorf("not enough devices for product %d (%s): need %d, found %d",
					item.ProductID, product.Name, item.Quantity, len(availableDevices))
			}

			// Add devices with discount
			for _, realDevice := range availableDevices {
				var discountedPrice *float64
				if product.ItemCostPerDay != nil {
					dp := *product.ItemCostPerDay * (1 - discountPercent)
					discountedPrice = &dp
				}

				packageIDInt := packageID
				realJobDevice := models.JobDevice{
					JobID:         jobID,
					DeviceID:      realDevice.DeviceID,
					CustomPrice:   discountedPrice,
					IsPackageItem: true,
					PackageID:     &packageIDInt,
				}

				if err := tx.Create(&realJobDevice).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to add real device %s: %w", realDevice.DeviceID, err)
				}

				log.Printf("[v4.0]   ↳ Added real device %s (%s) with discount: %.2f",
					realDevice.DeviceID, product.Name, *discountedPrice)
			}
		}
	}

	// Create job_package record for tracking (optional, for backwards compatibility)
	var priceValue sql.NullFloat64
	if customPrice != nil {
		priceValue = sql.NullFloat64{Float64: *customPrice, Valid: true}
	}

	jobPackage := &models.JobPackage{
		JobID:       jobID,
		PackageID:   packageID,
		Quantity:    quantity,
		CustomPrice: priceValue,
		AddedAt:     time.Now(),
		AddedBy:     &userID,
	}

	if err := tx.Create(jobPackage).Error; err != nil {
		// If already exists, just continue - idempotency
		if !strings.Contains(err.Error(), "Duplicate entry") {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create job_package record: %w", err)
		}
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Printf("=== AssignPackageToJob v4.0 SUCCESS: assigned %d package devices to job %d ===", quantity, jobID)

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
		SELECT COUNT(DISTINCT d.deviceID)
		FROM devices d
		WHERE d.productID = ?
		AND (
			EXISTS (
				SELECT 1 FROM jobdevices jd
				JOIN jobs j ON jd.jobID COLLATE utf8mb4_unicode_ci = j.jobID COLLATE utf8mb4_unicode_ci
				WHERE jd.deviceID COLLATE utf8mb4_unicode_ci = d.deviceID COLLATE utf8mb4_unicode_ci
				AND j.jobID != ?
				AND j.startDate IS NOT NULL
				AND j.endDate IS NOT NULL
				AND j.startDate <= ?
				AND j.endDate >= ?
			)
			OR EXISTS (
				SELECT 1 FROM job_package_reservations jpr
				JOIN job_packages jp ON jpr.job_package_id = jp.job_package_id
				JOIN jobs j ON jp.job_id = j.jobID
				WHERE jpr.device_id COLLATE utf8mb4_unicode_ci = d.deviceID COLLATE utf8mb4_unicode_ci
				AND jpr.reservation_status = 'reserved'
				AND j.jobID != ?
				AND j.startDate IS NOT NULL
				AND j.endDate IS NOT NULL
				AND j.startDate <= ?
				AND j.endDate >= ?
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
		SELECT d.deviceID
		FROM devices d
		WHERE d.productID = ?
		AND NOT EXISTS (
			SELECT 1 FROM jobdevices jd
			JOIN jobs j ON jd.jobID COLLATE utf8mb4_unicode_ci = j.jobID COLLATE utf8mb4_unicode_ci
			WHERE jd.deviceID COLLATE utf8mb4_unicode_ci = d.deviceID COLLATE utf8mb4_unicode_ci
			AND j.jobID != ?
			AND j.startDate IS NOT NULL
			AND j.endDate IS NOT NULL
			AND j.startDate <= ?
			AND j.endDate >= ?
		)
		AND NOT EXISTS (
			SELECT 1 FROM job_package_reservations jpr
			JOIN job_packages jp ON jpr.job_package_id = jp.job_package_id
			JOIN jobs j ON jp.job_id = j.jobID
			WHERE jpr.device_id COLLATE utf8mb4_unicode_ci = d.deviceID COLLATE utf8mb4_unicode_ci
			AND jpr.reservation_status = 'reserved'
			AND j.jobID != ?
			AND j.startDate IS NOT NULL
			AND j.endDate IS NOT NULL
			AND j.startDate <= ?
			AND j.endDate >= ?
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
	err := r.db.
		Preload("Package").
		Preload("Reservations").
		Preload("Reservations.Device").
		Preload("AddedByUser").
		Where("job_package_id = ?", id).
		First(&jobPackage).Error

	if err != nil {
		return nil, err
	}

	return &jobPackage, nil
}

// GetJobPackagesByJobID retrieves all packages for a job
func (r *JobPackageRepository) GetJobPackagesByJobID(jobID int) ([]models.JobPackage, error) {
	var packages []models.JobPackage
	err := r.db.
		Preload("Package").
		Preload("Reservations").
		Preload("Reservations.Device").
		Preload("AddedByUser").
		Where("job_id = ?", jobID).
		Order("added_at DESC").
		Find(&packages).Error

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

	// Update reservation status to released
	if err := tx.Model(&models.JobPackageReservation{}).
		Where("job_package_id = ?", jobPackageID).
		Updates(map[string]interface{}{
			"reservation_status": "released",
			"released_at":        time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return err
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
	return r.db.Model(&models.JobPackageReservation{}).
		Where("job_package_id = ? AND reservation_status = 'reserved'", jobPackageID).
		Updates(map[string]interface{}{
			"reservation_status": "released",
			"released_at":        time.Now(),
		}).Error
}

// GetJobPackagesWithDetails retrieves packages with computed details for display
func (r *JobPackageRepository) GetJobPackagesWithDetails(jobID int) ([]models.JobPackageWithDetails, error) {
	var packages []models.JobPackage
	err := r.db.
		Preload("Package").
		Preload("Reservations").
		Where("job_id = ?", jobID).
		Order("added_at DESC").
		Find(&packages).Error

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
			var itemCount int64
			r.db.Model(&models.ProductPackageItem{}).Where("package_id = ?", pkg.Package.PackageID).Count(&itemCount)
			details.DeviceCount = int(itemCount)
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
