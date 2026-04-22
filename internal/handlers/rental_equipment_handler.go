package handlers

import (
	"net/http"
	"strconv"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RentalEquipmentHandler struct {
	repo *repository.RentalEquipmentRepository
}

func NewRentalEquipmentHandler(repo *repository.RentalEquipmentRepository) *RentalEquipmentHandler {
	return &RentalEquipmentHandler{repo: repo}
}

func bindingErrorsToMap(err error) map[string]string {
	fields := map[string]string{}
	if err == nil {
		return fields
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			fields[e.Field()] = e.Error()
		}
		return fields
	}
	// fallback: single error
	fields["_error"] = err.Error()
	return fields
}

// ShowRentalEquipmentList renders a deprecation notice for rental equipment management
func (h *RentalEquipmentHandler) ShowRentalEquipmentList(c *gin.Context) {
	user, _ := GetCurrentUser(c)

	c.HTML(http.StatusOK, "rental_equipment_standalone.html", gin.H{
		"title":       "Rental Equipment",
		"user":        user,
		"currentPage": "rental-equipment",
	})
}

// ShowRentalEquipmentForm renders a deprecation notice for the legacy form
func (h *RentalEquipmentHandler) ShowRentalEquipmentForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)

	c.HTML(http.StatusOK, "rental_equipment_form_standalone.html", gin.H{
		"title":       "Rental Equipment",
		"user":        user,
		"currentPage": "rental-equipment",
	})
}

// ShowRentalAnalytics renders a deprecation notice for legacy analytics
func (h *RentalEquipmentHandler) ShowRentalAnalytics(c *gin.Context) {
	user, _ := GetCurrentUser(c)

	c.HTML(http.StatusOK, "rental_equipment_analytics_standalone.html", gin.H{
		"title":       "Rental Equipment Analytics",
		"user":        user,
		"currentPage": "rental-analytics",
	})
}

func (h *RentalEquipmentHandler) CreateRentalEquipment(c *gin.Context) {
	// create rental equipment (legacy UI) - use CreateRentalEquipmentRequest
	var req models.CreateRentalEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": bindingErrorsToMap(err)})
		return
	}
	user, _ := GetCurrentUser(c)
	var createdBy *uint
	if user != nil {
		createdBy = &user.UserID
	}
	equipment := &models.RentalEquipment{
		ProductName:  req.ProductName,
		SupplierName: req.SupplierName,
		RentalPrice:  req.RentalPrice,
		Category:     req.Category,
		Description:  req.Description,
		Notes:        req.Notes,
		IsActive:     req.IsActive,
		CreatedBy:    createdBy,
	}
	if err := h.repo.CreateRentalEquipment(equipment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rental equipment"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"rentalEquipment": equipment})
}

func (h *RentalEquipmentHandler) UpdateRentalEquipment(c *gin.Context) {
	// update rental equipment
	var req models.UpdateRentalEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": bindingErrorsToMap(err)})
		return
	}
	// equipment id from URL
	idParam := c.Param("id")
	// parse uint
	// reuse models.RentalEquipment struct
	var existing models.RentalEquipment
	// convert id
	// simple Atoi
	// avoid extra import by using Gin's Param and ParseUint
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}
	// parse
	var equipmentID uint64
	if n, err := strconv.ParseUint(idParam, 10, 64); err == nil {
		equipmentID = n
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.repo.GetRentalEquipmentByID(uint(equipmentID), &existing); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rental equipment not found"})
		return
	}
	existing.ProductName = req.ProductName
	existing.SupplierName = req.SupplierName
	existing.RentalPrice = req.RentalPrice
	existing.Category = req.Category
	existing.Description = req.Description
	existing.Notes = req.Notes
	existing.IsActive = req.IsActive

	if err := h.repo.UpdateRentalEquipment(&existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rentalEquipment": existing})
}

func (h *RentalEquipmentHandler) DeleteRentalEquipment(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.repo.DeleteRentalEquipment(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// GetRentalEquipmentAPI godoc
// @Summary      Rental equipment API deprecated
// @Description  This endpoint has moved to WarehouseCore and always returns HTTP 410 Gone.
// @Tags         rental-equipment
// @Produce      json
// @Failure      410  {object}  map[string]string  "Feature moved to WarehouseCore"
// @Security     SessionCookie
// @Router       /rental-equipment [get]
func (h *RentalEquipmentHandler) GetRentalEquipmentAPI(c *gin.Context) {
	var items []models.RentalEquipment
	if err := h.repo.GetAllRentalEquipment(&items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list rental equipment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rentalEquipment": items})
}

func (h *RentalEquipmentHandler) AddRentalToJob(c *gin.Context) {
	var req models.AddRentalToJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": bindingErrorsToMap(err)})
		return
	}
	jobRental := &models.JobRentalEquipment{
		JobID:       req.JobID,
		EquipmentID: req.EquipmentID,
		Quantity:    req.Quantity,
		DaysUsed:    req.DaysUsed,
		Notes:       req.Notes,
	}
	if err := h.repo.AddRentalToJob(jobRental); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add rental to job", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobRental": jobRental})
}

func (h *RentalEquipmentHandler) CreateManualRentalEntry(c *gin.Context) {
	var req models.ManualRentalEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": bindingErrorsToMap(err)})
		return
	}
	user, _ := GetCurrentUser(c)
	var createdBy *uint
	if user != nil {
		createdBy = &user.UserID
	}
	equipment, jobRental, err := h.repo.CreateRentalEquipmentFromManualEntry(&req, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create manual rental entry", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"rentalEquipment": equipment, "jobRental": jobRental})
}

func (h *RentalEquipmentHandler) GetJobRentalEquipment(c *gin.Context) {
	jobIdParam := c.Param("jobId")
	if jobIdParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing job id"})
		return
	}
	id, err := strconv.ParseUint(jobIdParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	var items []models.JobRentalEquipment
	if err := h.repo.GetJobRentalEquipment(uint(id), &items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch job rentals"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobRentals": items})
}

func (h *RentalEquipmentHandler) RemoveRentalFromJob(c *gin.Context) {
	jobIdParam := c.Param("jobId")
	eqParam := c.Param("equipmentId")
	if jobIdParam == "" || eqParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameters"})
		return
	}
	jobID, err := strconv.ParseUint(jobIdParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	eqID, err := strconv.ParseUint(eqParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid equipment id"})
		return
	}
	if err := h.repo.RemoveRentalFromJob(uint(jobID), uint(eqID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove rental from job", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"removed": true})
}

func (h *RentalEquipmentHandler) GetRentalAnalyticsAPI(c *gin.Context) {
	rentalEquipmentFeatureMovedJSON(c)
}

func rentalEquipmentFeatureMovedJSON(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"error":   "Rental equipment functionality has moved to WarehouseCore",
		"message": "Use WarehouseCore to manage rental equipment and analytics.",
	})
}
