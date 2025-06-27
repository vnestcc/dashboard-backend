package company

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils"
	"github.com/vnestcc/dashboard/utils/values"
	"gorm.io/gorm"
)

// ListQuater godoc
// @Summary      List quarters by company
// @Description  Lists all quarters for the specified company
// @Tags         company
// @Produce      json
// @Param        id   path      int  true  "Company ID"
// @Success      200  {array}   quarterResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /company/quarters/{id} [get]
func ListQuater(ctx *gin.Context) {
	var db = values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "list_quarter",
	})
	idStr := ctx.Param("id")
	companyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "invalid_company_id",
			"company_id": idStr,
		}).Warn("Invalid company ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var company models.Company
	if cached, found := StartupCache.Get(uint(companyID)); found {
		company = cached
	} else {
		if err := db.First(&company, uint(companyID)).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"status":     "failure",
				"reason":     "company_not_found",
				"company_id": companyID,
				"error":      err.Error(),
			}).Warn("Company not found")
			ctx.Set("message", err.Error())
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch company"})
			return
		}
		StartupCache.Set(company.ID, company)
	}
	if err := db.Model(&company).Association("Quarters").Find(&company.Quarters); err != nil {
		auditLog.WithFields(logrus.Fields{
			"status":     "failure",
			"reason":     "fetch_quarters_failed",
			"company_id": company.ID,
			"error":      err.Error(),
		}).Error("Failed to fetch quarters")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quarters"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"status":     "success",
		"company_id": company.ID,
		"quarters":   len(company.Quarters),
	}).Info("Fetched company quarters")
	ctx.JSON(http.StatusOK, company.Quarters)
}

// ListCompanyAdmin godoc
// @Summary      List all companies
// @Description  Retrieves a list of all companies available in the system
// @Tags         admin
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /manage/company/list [get]
func ListCompanyAdmin(ctx *gin.Context) {
	// placeholder func
}

// ListCompany godoc
// @Summary      List all companies
// @Description  Retrieves a list of all companies available in the system
// @Tags         company
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /company/list [get]
func ListCompany(ctx *gin.Context) {
	var db = values.GetDB()
	auditLog := utils.Logger.WithFields(logrus.Fields{
		"ip":    ctx.ClientIP(),
		"type":  "audit",
		"event": "list_company",
	})
	var companies []models.Company
	result := make(map[uint]string)
	if err := db.Find(&companies).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"status": "failure",
			"error":  err.Error(),
		}).Error("Failed to retrieve the company list")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve the company list"})
		return
	}
	for i := range companies {
		StartupCache.Set(companies[i].ID, companies[i])
		result[companies[i].ID] = companies[i].Name
	}
	auditLog.WithFields(logrus.Fields{
		"status":        "success",
		"company_count": len(result),
	}).Info("Fetched company list")
	ctx.JSON(http.StatusOK, result)
}
