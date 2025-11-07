package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/services/pdf"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PDFHandler handles PDF upload and processing requests
type PDFHandler struct {
	DB        *gorm.DB
	Extractor *pdf.PDFExtractor
	Mapper    *pdf.ProductMapper
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(db *gorm.DB, uploadDir string) *PDFHandler {
	return &PDFHandler{
		DB:        db,
		Extractor: pdf.NewPDFExtractor(uploadDir),
		Mapper:    pdf.NewProductMapper(db),
	}
}

// UploadPDF handles PDF file upload
// POST /api/v1/pdf/upload
func (h *PDFHandler) UploadPDF(c *gin.Context) {
	// Get job ID if provided
	jobIDStr := c.PostForm("job_id")
	var jobID sql.NullInt64
	if jobIDStr != "" {
		id, err := strconv.ParseInt(jobIDStr, 10, 64)
		if err == nil {
			jobID = sql.NullInt64{Int64: id, Valid: true}
		}
	}

	// Get uploaded file
	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file type
	if file.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
		return
	}

	// Save file
	upload, err := h.Extractor.SaveUploadedFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file: %v", err)})
		return
	}

	// Set job ID if provided
	upload.JobID = jobID

	// Get user ID from session
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(int64); ok {
			upload.UploadedBy = sql.NullInt64{Int64: uid, Valid: true}
		}
	}

	// Save upload record to database
	if err := h.DB.Create(upload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save upload record"})
		return
	}

	// Start processing asynchronously
	go h.processUploadAsync(upload.UploadID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"upload_id": upload.UploadID,
		"message":   "PDF uploaded successfully, processing started",
	})
}

// processUploadAsync processes the PDF asynchronously
func (h *PDFHandler) processUploadAsync(uploadID uint64) {
	// Update status to processing
	h.DB.Model(&models.PDFUpload{}).Where("upload_id = ?", uploadID).Updates(map[string]interface{}{
		"processing_status":    "processing",
		"processing_started_at": time.Now(),
	})

	// Get upload record
	var upload models.PDFUpload
	if err := h.DB.First(&upload, uploadID).Error; err != nil {
		h.markProcessingFailed(uploadID, fmt.Sprintf("Upload not found: %v", err))
		return
	}

	// Extract text
	rawText, err := h.Extractor.ExtractText(upload.FilePath)
	if err != nil {
		h.markProcessingFailed(uploadID, fmt.Sprintf("Text extraction failed: %v", err))
		return
	}

	// Parse invoice data
	parsedData, err := h.Extractor.ParseInvoiceData(rawText)
	if err != nil {
		h.markProcessingFailed(uploadID, fmt.Sprintf("Data parsing failed: %v", err))
		return
	}

	parsedData.RawText = rawText

	// Convert to JSON
	extractedDataJSON, err := parsedData.ToJSON()
	if err != nil {
		h.markProcessingFailed(uploadID, fmt.Sprintf("JSON conversion failed: %v", err))
		return
	}

	// Create extraction record
	extraction := models.PDFExtraction{
		UploadID:         uploadID,
		RawText:          sql.NullString{String: rawText, Valid: true},
		ExtractedData:    sql.NullString{String: extractedDataJSON, Valid: true},
		ConfidenceScore:  sql.NullFloat64{Float64: parsedData.ConfidenceScore, Valid: true},
		PageCount:        1, // TODO: Get actual page count
		ExtractionMethod: "regex_parser",
		CustomerName:     sql.NullString{String: parsedData.CustomerName, Valid: parsedData.CustomerName != ""},
		DocumentNumber:   sql.NullString{String: parsedData.DocumentNumber, Valid: parsedData.DocumentNumber != ""},
		TotalAmount:      sql.NullFloat64{Float64: parsedData.TotalAmount, Valid: parsedData.TotalAmount > 0},
		DiscountAmount:   sql.NullFloat64{Float64: parsedData.DiscountAmount, Valid: parsedData.DiscountAmount > 0},
	}

	if !parsedData.DocumentDate.IsZero() {
		extraction.DocumentDate = sql.NullTime{Time: parsedData.DocumentDate, Valid: true}
	}

	// Save extraction
	if err := h.DB.Create(&extraction).Error; err != nil {
		h.markProcessingFailed(uploadID, fmt.Sprintf("Failed to save extraction: %v", err))
		return
	}

	// Create extraction items
	for _, item := range parsedData.Items {
		extractionItem := models.PDFExtractionItem{
			ExtractionID:   extraction.ExtractionID,
			LineNumber:     sql.NullInt64{Int64: int64(item.LineNumber), Valid: true},
			RawProductText: item.ProductText,
			Quantity:       sql.NullInt64{Int64: int64(item.Quantity), Valid: true},
			UnitPrice:      sql.NullFloat64{Float64: item.UnitPrice, Valid: item.UnitPrice > 0},
			LineTotal:      sql.NullFloat64{Float64: item.LineTotal, Valid: item.LineTotal > 0},
			MappingStatus:  "pending",
		}

		// Try to find product mapping
		_, product, confidence, err := h.Mapper.FindBestMatch(item.ProductText)
		if err == nil && product != nil && confidence >= 70 {
			extractionItem.MappedProductID = sql.NullInt64{Int64: int64(product.ProductID), Valid: true}
			extractionItem.MappingConfidence = sql.NullFloat64{Float64: confidence, Valid: true}
			extractionItem.MappingStatus = "auto_mapped"
		}

		h.DB.Create(&extractionItem)
	}

	// Mark as completed
	h.DB.Model(&models.PDFUpload{}).Where("upload_id = ?", uploadID).Updates(map[string]interface{}{
		"processing_status":      "completed",
		"processing_completed_at": time.Now(),
	})
}

// markProcessingFailed marks upload as failed
func (h *PDFHandler) markProcessingFailed(uploadID uint64, errorMsg string) {
	h.DB.Model(&models.PDFUpload{}).Where("upload_id = ?", uploadID).Updates(map[string]interface{}{
		"processing_status":      "failed",
		"processing_completed_at": time.Now(),
		"error_message":          errorMsg,
	})
}

// GetExtractionResult retrieves extraction results
// GET /api/v1/pdf/extraction/:upload_id
func (h *PDFHandler) GetExtractionResult(c *gin.Context) {
	uploadID := c.Param("upload_id")

	var upload models.PDFUpload
	if err := h.DB.First(&upload, uploadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload not found"})
		return
	}

	var extraction models.PDFExtraction
	if err := h.DB.Where("upload_id = ?", uploadID).First(&extraction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Extraction not found"})
		return
	}

	var items []models.PDFExtractionItem
	h.DB.Where("extraction_id = ?", extraction.ExtractionID).Find(&items)

	// Build response
	response := models.PDFExtractionResponse{
		UploadID:     upload.UploadID,
		ExtractionID: extraction.ExtractionID,
		Items:        items,
	}

	if extraction.CustomerName.Valid {
		response.CustomerName = extraction.CustomerName.String
	}
	if extraction.CustomerID.Valid {
		customerID := int(extraction.CustomerID.Int64)
		response.CustomerID = &customerID
	}
	if extraction.DocumentNumber.Valid {
		response.DocumentNumber = extraction.DocumentNumber.String
	}
	if extraction.DocumentDate.Valid {
		response.DocumentDate = extraction.DocumentDate.Time.Format("2006-01-02")
	}
	if extraction.TotalAmount.Valid {
		response.TotalAmount = extraction.TotalAmount.Float64
	}
	if extraction.DiscountAmount.Valid {
		response.DiscountAmount = extraction.DiscountAmount.Float64
	}
	if extraction.RawText.Valid {
		response.RawText = extraction.RawText.String
	}
	if extraction.ConfidenceScore.Valid {
		response.ConfidenceScore = extraction.ConfidenceScore.Float64
	}

	c.JSON(http.StatusOK, response)
}

// SaveProductMapping saves a manual product mapping
// POST /api/v1/pdf/mapping
func (h *PDFHandler) SaveProductMapping(c *gin.Context) {
	var req struct {
		PDFText   string `json:"pdf_text" binding:"required"`
		ProductID int    `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := int64(1) // TODO: Get from session
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	if err := h.Mapper.SaveMapping(req.PDFText, req.ProductID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mapping saved successfully"})
}

// GetProductSuggestions gets product suggestions for PDF text
// GET /api/v1/pdf/suggestions?text=...
func (h *PDFHandler) GetProductSuggestions(c *gin.Context) {
	productText := c.Query("text")
	if productText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text parameter required"})
		return
	}

	suggestions, err := h.Mapper.FindSimilarProducts(productText, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

// UpdateItemMapping updates the product mapping for an extraction item
// PUT /api/v1/pdf/items/:item_id/mapping
func (h *PDFHandler) UpdateItemMapping(c *gin.Context) {
	itemID := c.Param("item_id")

	var req struct {
		ProductID int    `json:"product_id" binding:"required"`
		Status    string `json:"status"` // 'user_confirmed', 'user_rejected', etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "user_confirmed"
	}

	updates := map[string]interface{}{
		"mapped_product_id":  req.ProductID,
		"mapping_status":     status,
		"mapping_confidence": 100.0, // User confirmed = 100%
	}

	if err := h.DB.Model(&models.PDFExtractionItem{}).Where("item_id = ?", itemID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mapping updated successfully"})
}
