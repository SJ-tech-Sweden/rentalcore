package handlers

import (
	"log"
	"net/http"
	"strings"

	"go-barcode-webapp/internal/services"

	"github.com/gin-gonic/gin"
)

// TwentyHandler handles the Twenty CRM integration settings pages and API.
type TwentyHandler struct {
	twentyService *services.TwentyService
}

// NewTwentyHandler creates a new TwentyHandler.
func NewTwentyHandler(twentyService *services.TwentyService) *TwentyHandler {
	return &TwentyHandler{twentyService: twentyService}
}

// TwentySettingsForm renders the Twenty CRM integration settings page.
func (h *TwentyHandler) TwentySettingsForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	cfg := h.twentyService.GetConfig()

	var successMsg string
	if c.Query("success") == "1" {
		successMsg = "Twenty integration settings saved successfully!"
	}

	c.HTML(http.StatusOK, "twenty_settings.html", gin.H{
		"title":       "Twenty CRM Integration",
		"user":        user,
		"config":      cfg,
		"success":     successMsg,
		"currentPage": "settings",
	})
}

// twentySettingsRequest is the request body for updating Twenty integration settings.
type twentySettingsRequest struct {
	Enabled bool   `json:"enabled"`
	APIURL  string `json:"apiUrl"`
	APIKey  string `json:"apiKey"`
}

// twentySettingsResponse is the response body for Twenty integration settings.
type twentySettingsResponse struct {
	Success bool   `json:"success"`
	Enabled bool   `json:"enabled"`
	APIURL  string `json:"apiUrl"`
	// APIKey is intentionally omitted from read responses for security.
}

// GetTwentySettings returns the current Twenty integration configuration.
//
// @Summary     Get Twenty CRM integration settings
// @Description Returns the current configuration for the Twenty CRM integration.
// @Tags        admin
// @Produce     json
// @Success     200 {object} twentySettingsResponse
// @Router      /api/v1/admin/integrations/twenty [get]
// @Security    SessionCookie
func (h *TwentyHandler) GetTwentySettings(c *gin.Context) {
	cfg := h.twentyService.GetConfig()
	c.JSON(http.StatusOK, twentySettingsResponse{
		Success: true,
		Enabled: cfg.Enabled,
		APIURL:  cfg.APIURL,
	})
}

// UpdateTwentySettings saves updated Twenty integration settings.
//
// @Summary     Update Twenty CRM integration settings
// @Description Saves the Twenty CRM integration configuration (API URL, API key, enabled flag).
// @Tags        admin
// @Accept      json
// @Produce     json
// @Param       body body twentySettingsRequest true "Twenty settings"
// @Success     200 {object} twentySettingsResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Router      /api/v1/admin/integrations/twenty [put]
// @Security    SessionCookie
func (h *TwentyHandler) UpdateTwentySettings(c *gin.Context) {
	var req twentySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	req.APIURL = strings.TrimSpace(req.APIURL)
	req.APIKey = strings.TrimSpace(req.APIKey)

	if req.Enabled && req.APIURL == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "API URL is required when enabling the Twenty integration"})
		return
	}
	if req.Enabled && req.APIKey == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "API key is required when enabling the Twenty integration"})
		return
	}

	cfg := services.TwentyConfig{
		Enabled: req.Enabled,
		APIURL:  req.APIURL,
		APIKey:  req.APIKey,
	}
	if err := h.twentyService.SaveConfig(cfg); err != nil {
		log.Printf("UpdateTwentySettings: failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "failed to save Twenty settings"})
		return
	}
	c.JSON(http.StatusOK, twentySettingsResponse{
		Success: true,
		Enabled: cfg.Enabled,
		APIURL:  cfg.APIURL,
	})
}

// TestTwentyConnection tests connectivity to the configured Twenty CRM instance.
//
// @Summary     Test Twenty CRM connection
// @Description Attempts to connect to the Twenty CRM API and returns the result.
// @Tags        admin
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} errorResponse
// @Router      /api/v1/admin/integrations/twenty/test [post]
// @Security    SessionCookie
func (h *TwentyHandler) TestTwentyConnection(c *gin.Context) {
	if err := h.twentyService.TestConnection(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection to Twenty CRM successful",
	})
}
