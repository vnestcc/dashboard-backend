package company

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils"
	"github.com/vnestcc/dashboard/utils/values"
)

// GetCompanyByIDAdmin godoc
// @Summary      Get company details (Admin)
// @Description  Returns the specified company's information, including selectable related data sets (admin only).
// @Tags         admin
// @Produce      json
// @Param        id       path   int     true  "Company ID"
// @Param        data     query  string  false "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)
// @Param        quarter  query  string  false "Quarter (e.g. Q1, Q2, Q3, Q4)"
// @Param        year     query  int     false "Year"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/{id} [get]
func GetCompanyByIDAdmin(ctx *gin.Context) {
	// dead function just a place holder for docs
}

// EditCompanyByID godoc
// @Summary      Edit company details (Admin, versioned insert)
// @Description  Allows admin to insert new versioned data for company or related quarter data. If `data=info` or omitted, updates company name, contact name, and contact email. Otherwise, allows versioned updates for specific company data types (such as finance, market, uniteconomics, etc) for a given quarter and year. The allowed types are: finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements. All data modifications will insert a new version for the specified quarter and year.
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "Company ID"
// @Param        data    query     string  false "Type of company data to edit (info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)"
// @Param        quarter query     string  false "Quarter name (e.g. Q1, Q2, Q3, Q4). Required unless data=info" Enum(Q1,Q2,Q3,Q4)
// @Param        year    query     int     false "Year (e.g. 2024). Required unless data=info"
// @Param        body    body      object  true  "Payload matching the type of data being edited"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string  "Invalid request or body"
// @Failure      404  {object}  map[string]string  "Company or quarter not found"
// @Failure      500  {object}  map[string]string  "Server/database error"
// @Router       /manage/company/edit/{id} [put]
// OPTIMIZE: whatever is written here is not meant for production. I pray for the server :pray:
// NOTE: need auditing
func EditCompanyByID(ctx *gin.Context) {
	var db = values.GetDB()
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)

	var company models.Company
	if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
		return
	}

	dataVals := ctx.Request.URL.Query()["data"]
	if len(dataVals) > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Only one 'data' query parameter allowed"})
		return
	}
	data := ""
	if len(dataVals) == 1 {
		data = dataVals[0]
	}
	quarter := ctx.Query("quarter")
	yearStr := ctx.Query("year")
	allowedData := map[string]string{
		"info":          "",
		"finance":       "FinancialHealths",
		"market":        "MarketTractions",
		"uniteconomics": "UnitEconomics",
		"teamperf":      "TeamPerformances",
		"fund":          "FundraisingStatuses",
		"competitive":   "CompetitiveLandscapes",
		"operation":     "OperationalEfficiencies",
		"risk":          "RiskManagements",
		"additional":    "AdditionalInfos",
		"self":          "SelfAssessments",
		"attachements":  "Attachments",
	}
	if data != "" {
		if _, ok := allowedData[data]; !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
			return
		}
	}
	yearUint, err := strconv.ParseUint(yearStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	year := uint(yearUint)
	if data == "" || data == "info" {
		var req struct {
			Name         *string `json:"name"`
			ContactName  *string `json:"contact_name"`
			ContactEmail *string `json:"contact_email"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if req.Name != nil {
			company.Name = *req.Name
		}
		if req.ContactName != nil {
			company.ContactName = *req.ContactName
		}
		if req.ContactEmail != nil {
			company.ContactEmail = *req.ContactEmail
		}
		if err := db.Save(&company).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Company updated successfully", "company": company})
		return
	}
	var quarterObj models.Quarter
	if err := db.Where("company_id = ? AND quarter = ? AND year = ?", companyID, quarter, year).
		First(&quarterObj).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Quarter not found"})
		return
	}
	preloadField := allowedData[data]
	if preloadField == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No editable data specified"})
		return
	}
	switch data {
	case "finance":
		var req []models.FinancialHealth
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.FinancialHealth{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "FinancialHealths versioned and inserted"})
	case "market":
		var req []models.MarketTraction
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.MarketTraction{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "MarketTractions versioned and inserted"})
	case "uniteconomics":
		var req []models.UnitEconomics
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.UnitEconomics{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "UnitEconomics versioned and inserted"})
	case "teamperf":
		var req []models.TeamPerformance
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.TeamPerformance{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "TeamPerformances versioned and inserted"})
	case "fund":
		var req []models.FundraisingStatus
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.FundraisingStatus{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "FundraisingStatuses versioned and inserted"})
	case "competitive":
		var req []models.CompetitiveLandscape
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.CompetitiveLandscape{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "CompetitiveLandscapes versioned and inserted"})
	case "operation":
		var req []models.OperationalEfficiency
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.OperationalEfficiency{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "OperationalEfficiencies versioned and inserted"})
	case "risk":
		var req []models.RiskManagement
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.RiskManagement{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "RiskManagements versioned and inserted"})
	case "additional":
		var req []models.AdditionalInfo
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.AdditionalInfo{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "AdditionalInfos versioned and inserted"})
	case "self":
		var req []models.SelfAssessment
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.SelfAssessment{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "SelfAssessments versioned and inserted"})
	case "attachements":
		var req []models.Attachment
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		for _, newObj := range req {
			newObj.QuarterID = quarterObj.ID
			var maxVersion int
			db.Model(&models.Attachment{}).
				Where("quarter_id = ?", quarterObj.ID).
				Select("COALESCE(MAX(version),0)").Scan(&maxVersion)
			newObj.Version = uint32(1)
			db.Create(&newObj)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Attachments versioned and inserted"})
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
	}
}

// DeleteCompanyByID godoc
// @Summary      Admin delete company
// @Description  Deletes a company by its ID (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Company ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /manage/company/delete/{id} [delete]
func DeleteCompanyByID(ctx *gin.Context) {
	var db = values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "delete_company_by_id",
	})
	idStr := ctx.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "invalid_company_id",
			"company_id": idStr,
		}).Warn("Invalid company ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(idUint)
	tx := db.Delete(&models.Company{}, companyID)
	if tx.Error != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "company_delete_failed",
			"company_id": companyID,
			"error":      tx.Error.Error(),
		}).Error("Failed to delete company")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}
	if tx.RowsAffected == 0 {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "company_not_exist",
			"company_id": companyID,
		}).Warn("Company does not exist")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Company does not exist"})
		return
	}
	if err := db.Model(&models.User{}).
		Where("startup_id = ?", companyID).
		Update("startup_id", nil).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "unlink_users_failed",
			"company_id": companyID,
			"error":      err.Error(),
		}).Error("Failed to clear company from users")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear company from users"})
		return
	}
	StartupCache.Delete(companyID)
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": companyID,
	}).Info("Company deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}
