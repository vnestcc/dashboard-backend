package company

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/handlers"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils"
	"github.com/vnestcc/dashboard/utils/values"
	"gorm.io/gorm"
)

// CreateCompany godoc
// @Summary      Create a company
// @Description  Creates a new company for the user
// @Tags         company
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body      createCompanyRequest true  "Company details"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/create [post]
func CreateCompany(ctx *gin.Context) {
	var db = values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "create_company",
	})

	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithField("status", "failure").Warn("Unauthorized: no claims in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok || claims.Role == "admin" {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"role":   claims.Role,
		}).Warn("Admins cannot create companies")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admins cannot create companies"})
		return
	}
	var req createCompanyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"reason":  "invalid_input",
			"details": err.Error(),
		}).Warn("Invalid input")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}
	var user models.User
	if value, ok := handlers.UserCache.Get(claims.ID); ok {
		user = value
	} else {
		if err := db.First(&user, claims.ID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"status":  "failure",
				"user_id": claims.ID,
				"error":   err.Error(),
			}).Error("Failed to fetch user")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		}
		handlers.UserCache.Set(claims.ID, user)
	}
	if user.StartupID != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": user.ID,
			"reason":  "already_in_company",
		}).Warn("User already belongs to a company")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "User already belongs to a company"})
		return
	}
	newCompany := models.Company{
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		Sector:       req.Sector,
		Description:  req.Description,
	}
	if err := db.Create(&newCompany).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			auditLog.WithFields(logrus.Fields{
				"status": "failure",
				"reason": "duplicate_email",
				"email":  req.ContactEmail,
				"error":  err.Error(),
			}).Warn("Company with this contact email already exists")
			ctx.JSON(http.StatusConflict, gin.H{"error": "Company with this contact email already exists"})
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "db_error",
			"error":  err.Error(),
		}).Error("Failed to create company")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}
	if err := db.Model(&user).Update("startup_id", newCompany.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"user_id":    user.ID,
			"company_id": newCompany.ID,
			"error":      err.Error(),
		}).Error("Failed to link company to user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link company to user"})
		return
	}
	StartupCache.Set(newCompany.ID, newCompany)
	handlers.UserCache.Set(user.ID, user)
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": newCompany.ID,
		"user_id":    user.ID,
	}).Info("Company created and linked to user")
	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Company created successfully",
		"company_id": newCompany.ID,
	})
}

// UserCompany godoc
// @Summary      Get current user's company
// @Description  Retrieves the company associated with the authenticated user
// @Tags         company
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /company/me [get]
func UserCompany(ctx *gin.Context) {
	var db = values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "user_company",
	})
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithField("status", "failure").Warn("Unauthorized access: no claims in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		auditLog.WithField("status", "failure").Warn("Invalid claims format")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if err := db.Preload("StartUp").First(&user, claims.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": claims.ID,
			"error":   err.Error(),
		}).Error("Failed to fetch user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user.StartUp == nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": claims.ID,
		}).Warn("User does not belong to any company")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User does not belong to any company"})
		return
	}
	startup := user.StartUp
	if _, ok := StartupCache.Get(startup.ID); !ok {
		StartupCache.Set(startup.ID, *startup)
	}
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"user_id":    claims.ID,
		"company_id": startup.ID,
	}).Info("Fetched user's company")
	ctx.JSON(http.StatusOK, gin.H{
		"id":            startup.ID,
		"name":          startup.Name,
		"secret_code":   startup.SecretCode,
		"contact_name":  startup.ContactName,
		"contact_email": startup.ContactEmail,
	})
}

// DeleteCompany godoc
// @Summary      Delete a company
// @Description  Deletes the user's company
// @Tags         company
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /company/delete [delete]
func DeleteCompany(ctx *gin.Context) {
	var db = values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "delete_company",
	})
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithField("status", "failure").Warn("Unauthorized: no claims in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		auditLog.WithField("status", "failure").Warn("Invalid claims format")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if err := db.Preload("StartUp").First(&user, claims.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": claims.ID,
			"error":   err.Error(),
		}).Error("Failed to fetch user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user.StartUp == nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": user.ID,
			"reason":  "no_company",
		}).Warn("User does not belong to any company")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User does not belong to any company"})
		return
	}
	companyID := user.StartUp.ID
	if err := db.Delete(&models.Company{}, companyID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"user_id":    user.ID,
			"company_id": companyID,
			"error":      err.Error(),
		}).Error("Failed to delete company")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}
	if err := db.Model(&models.User{}).
		Where("startup_id = ?", companyID).
		Update("startup_id", nil).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"user_id":    user.ID,
			"company_id": companyID,
			"error":      err.Error(),
		}).Error("Failed to clear company from users")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear company from users"})
		return
	}
	StartupCache.Delete(companyID)
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"user_id":    user.ID,
		"company_id": companyID,
	}).Info("Company deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}

// JoinCompany godoc
// @Summary      Join a company
// @Description  Join company for the user
// @Tags         company
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      int                true  "Company ID"
// @Param        body body      joinCompanyRequest true  "Secret Code"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/join/{id} [post]
func JoinCompany(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "join_company",
	})
	companyIDStr := ctx.Param("id")
	companyIDUint, err := strconv.ParseUint(companyIDStr, 10, 32)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "invalid_company_id",
			"company_id": companyIDStr,
		}).Warn("Invalid company ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	companyID := uint(companyIDUint)
	var req struct {
		SecretCode string `json:"secret_code" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "missing_or_invalid_secret_code",
		}).Warn("Missing or invalid secret code")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid secret code"})
		return
	}
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithField("status", "failure").Warn("Unauthorized: no claims in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		auditLog.WithField("status", "failure").Warn("Invalid claims format")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	userID := claims.ID
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to retrieve user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	if user.StartupID != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": user.ID,
			"reason":  "already_in_company",
		}).Warn("User already belongs to a company")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "User already belongs to a company"})
		return
	}
	var company models.Company
	if cachedCompany, found := StartupCache.Get(companyID); found {
		company = cachedCompany
	} else {
		if err := db.First(&company, companyID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				auditLog.WithFields(logrus.Fields{
					"status":     "failure",
					"reason":     "company_not_found",
					"company_id": companyID,
				}).Warn("Company not found")
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"reason":     "company_fetch_failed",
				"company_id": companyID,
				"error":      err.Error(),
			}).Error("Failed to retrieve company")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve company"})
			return
		}
		StartupCache.Set(company.ID, company)
	}
	if company.SecretCode != req.SecretCode {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "invalid_secret_code",
			"company_id": companyID,
			"user_id":    userID,
		}).Warn("Invalid secret code")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid secret code"})
		return
	}
	if err := db.Model(&user).Update("startup_id", company.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "db_error",
			"user_id":    user.ID,
			"company_id": company.ID,
			"error":      err.Error(),
		}).Error("Failed to join company")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join company"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": company.ID,
		"user_id":    user.ID,
	}).Info("Successfully joined the company")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully joined the company"})
}

func handleEdit[T editableModel](
	ctx *gin.Context,
	db *gorm.DB,
	quarterObj *models.Quarter,
	table string,
	auditLog *logrus.Entry,
) {
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "invalid_request_body",
			"table":  table,
		}).Warn("Invalid request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	v := reflect.ValueOf(req)
	var version uint32
	var isEditable uint16
	var model T

	row := db.
		Model(&model).
		Select("version, is_editable").
		Where("quarter_id = ? AND company_id = ?", quarterObj.ID, quarterObj.CompanyID).
		Order("version DESC").
		Limit(1).
		Row()
	err := row.Scan(&version, &isEditable)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			version = 1
			isEditable = 1023
		} else {
			auditLog.WithFields(logrus.Fields{
				"status":  "failure",
				"error":   "db_scan_error",
				"table":   table,
				"details": err.Error(),
			}).Error("Failed to fetch latest version")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch latest version"})
			return
		}
	} else {
		version = version + uint32(1)
	}
	v.Elem().FieldByName("Version").SetUint(uint64(version))
	v.Elem().FieldByName("IsEditable").SetUint(uint64(isEditable))
	if err := req.EditableFilter(); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "edit_mask_restricted",
			"table":  table,
		}).Warn("Edit not permitted by field-level mask")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	v.Elem().FieldByName("QuarterID").SetUint(uint64(quarterObj.ID))
	v.Elem().FieldByName("CompanyID").SetUint(uint64(quarterObj.CompanyID))
	if err := db.Create(&req).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"error":   "db_create_failed",
			"table":   table,
			"details": err.Error(),
		}).Error("Failed to insert record")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add %s data", table)})
		return
	}

	getID := func() uint {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		idField := v.FieldByName("ID")
		if idField.IsValid() && idField.CanUint() {
			return uint(idField.Uint())
		}
		return 0
	}

	auditLog.WithFields(logrus.Fields{
		"status":  "success",
		"table":   table,
		"version": version,
		"id":      getID(),
	}).Info("Record versioned and inserted")

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s versioned and inserted", table),
		"version": version,
		"id":      getID(),
	})
}

// EditCompany godoc
// @Summary      Edit company information
// @Description  Updates the existing company data. If `data=info` or omitted, updates company name, contact name, and contact email. Otherwise, allows versioned updates for specific company data types (such as finance, market, uniteconomics, etc) for a given quarter and year. The allowed types are: finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements. All data modifications are subject to field-level editability checks based on the current IsEditable mask for each record.
// @Security     BearerAuth
// @Tags         company
// @Accept       json
// @Produce      json
// @Param        data   query     string  false  "Which related data to include"  Enums(info, finance, market, uniteconomics, teamperf, fund, competitive, operation, risk, additional, self, attachements)
// @Param        quarter query    string  false  "Quarter name (e.g. Q1, Q2, Q3, Q4). Required unless data=info" Enum(Q1,Q2,Q3,Q4)
// @Param        year    query    int  false  "Year (e.g. 2024). Required unless data=info"
// @Param        body    body     object  true   "Payload matching the type of data being edited"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string  "Invalid request or body"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Not permitted by edit mask"
// @Failure      404  {object}  map[string]string  "Company or quarter not found"
// @Failure      500  {object}  map[string]string  "Server/database error"
// @Router       /company/edit [put]
func EditCompany(ctx *gin.Context) {
	db := values.GetDB()
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}
	var user models.User
	if err := db.Preload("StartUp").First(&user, claims.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	companyID := user.StartupID
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"type":       "audit",
		"ip":         ctx.ClientIP(),
		"event":      "edit_company",
		"user_id":    user.ID,
		"company_id": companyID,
	})
	var company models.Company
	if err := db.Where("id = ?", companyID).First(&company).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  "company_not_found",
		}).Warn("Company not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not find company"})
		return
	}
	dataVals := ctx.Request.URL.Query()["data"]
	if len(dataVals) > 1 {
		auditLog.WithField("status", "failure").Warn("Multiple 'data' parameters")
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
	}
	if data != "" {
		if _, ok := allowedData[data]; !ok {
			auditLog.WithField("status", "failure").Warn("Invalid data type")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
			return
		}
	}
	yearUint, err := strconv.ParseUint(yearStr, 10, 32)
	if err != nil {
		auditLog.WithField("status", "failure").Warn("Invalid year format")
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
			auditLog.WithField("status", "failure").Warn("Invalid company info body")
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
			auditLog.WithFields(logrus.Fields{
				"status": "failure",
				"error":  err.Error(),
			}).Error("Failed to update company info")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
			return
		}
		auditLog.WithField("status", "success").Info("Updated company info")
		ctx.JSON(http.StatusOK, gin.H{"message": "Company updated successfully", "company": company})
		return
	}
	var quarterObj models.Quarter
	if err := db.Where("company_id = ? AND quarter = ? AND year = ?", companyID, quarter, year).
		First(&quarterObj).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"quarter": quarter,
			"year":    year,
		}).Warn("Quarter not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Quarter not found"})
		return
	}
	preloadField := allowedData[data]
	if preloadField == "" {
		auditLog.WithField("status", "failure").Warn("No editable data specified")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No editable data specified"})
		return
	}
	switch data {
	case "finance":
		handleEdit[*models.FinancialHealth](ctx, db, &quarterObj, preloadField, auditLog)
	case "market":
		handleEdit[*models.MarketTraction](ctx, db, &quarterObj, preloadField, auditLog)
	case "uniteconomics":
		handleEdit[*models.UnitEconomics](ctx, db, &quarterObj, preloadField, auditLog)
	case "teamperf":
		handleEdit[*models.TeamPerformance](ctx, db, &quarterObj, preloadField, auditLog)
	case "fund":
		handleEdit[*models.FundraisingStatus](ctx, db, &quarterObj, preloadField, auditLog)
	case "competitive":
		handleEdit[*models.CompetitiveLandscape](ctx, db, &quarterObj, preloadField, auditLog)
	case "operation":
		handleEdit[*models.OperationalEfficiency](ctx, db, &quarterObj, preloadField, auditLog)
	case "risk":
		handleEdit[*models.RiskManagement](ctx, db, &quarterObj, preloadField, auditLog)
	case "additional":
		handleEdit[*models.AdditionalInfo](ctx, db, &quarterObj, preloadField, auditLog)
	case "self":
		handleEdit[*models.SelfAssessment](ctx, db, &quarterObj, preloadField, auditLog)
	default:
		auditLog.WithField("status", "failure").Warn("Unexpected data type after validation")
		// assessment comes for upload
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query parameter"})
	}
}

// AddQuarter godoc
// @Summary      Add a new quarter
// @Description  Adds a new quarter record for the user's company
// @Tags         company
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body      quarterRequest true "Quarter details"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/quarters/add [post]
func AddQuarter(ctx *gin.Context) {
	db := values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "add_quarter",
	})
	claimsVal, exists := ctx.Get("claims")
	if !exists {
		auditLog.WithField("status", "failure").Warn("Unauthorized: no claims in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims, ok := claimsVal.(*Claims)
	if !ok || claims.Role == "admin" {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"role":   claims.Role,
		}).Warn("Admins cannot create quarters")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admins cannot create quarters"})
		return
	}
	var req quarterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"reason":  "invalid_input",
			"details": err.Error(),
		}).Warn("Invalid input")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}
	var user models.User
	if value, ok := handlers.UserCache.Get(claims.ID); ok {
		user = value
	} else {
		if err := db.First(&user, claims.ID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"status":  "failure",
				"user_id": claims.ID,
				"error":   err.Error(),
			}).Error("Failed to fetch user")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		}
		handlers.UserCache.Set(claims.ID, user)
	}
	if user.StartupID == nil {
		auditLog.WithFields(logrus.Fields{
			"status":  "failure",
			"user_id": user.ID,
			"reason":  "no_company",
		}).Warn("User does not belong to a company")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "User does not belong to a company"})
		return
	}
	newQuarter := models.Quarter{
		CompanyID: *user.StartupID,
		Quarter:   req.Quarter,
		Year:      req.Year,
		Date:      time.Now(),
	}
	if err := db.Create(&newQuarter).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"reason":     "duplicate_quarter",
				"company_id": *user.StartupID,
				"quarter":    req.Quarter,
				"year":       req.Year,
			}).Warn("Quarter already exists for company")
			ctx.JSON(http.StatusConflict, gin.H{"error": "Quarter already exists for this company and year"})
			return
		}
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"reason": "db_error",
			"error":  err.Error(),
		}).Error("Failed to create quarter")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quarter"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"quarter_id": newQuarter.ID,
		"user_id":    user.ID,
		"company_id": *user.StartupID,
	}).Info("Quarter created successfully")
	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Quarter created successfully",
		"quarter_id": newQuarter.ID,
	})
}
