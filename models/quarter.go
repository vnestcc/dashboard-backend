package models

import (
	"time"

	"gorm.io/gorm"
)

type Quarter struct {
	gorm.Model
	ID        uint `gorm:"primaryKey;autoIncrement;uniqueIndex:idx_quarter_comp"`
	CompanyID uint `gorm:"not null;uniqueIndex:idx_quarter_comp;uniqueIndex:idx_company_quarter_year"`
	Date      time.Time
	Quarter   string `gorm:"not null;uniqueIndex:idx_company_quarter_year"`
	Year      uint   `gorm:"not null;uniqueIndex:idx_company_quarter_year"`

	Company                 Company                 `gorm:"foreignKey:CompanyID"`
	FinancialHealths        []FinancialHealth       `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	MarketTractions         []MarketTraction        `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	UnitEconomics           []UnitEconomics         `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	ProductDevelopments     []ProductDevelopment    `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	TeamPerformances        []TeamPerformance       `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	FundraisingStatuses     []FundraisingStatus     `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	CompetitiveLandscapes   []CompetitiveLandscape  `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	OperationalEfficiencies []OperationalEfficiency `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	RiskManagements         []RiskManagement        `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	AdditionalInfos         []AdditionalInfo        `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	SelfAssessments         []SelfAssessment        `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
	Attachments             []Attachment            `gorm:"foreignKey:QuarterID,CompanyID;references:ID,CompanyID"`
}
