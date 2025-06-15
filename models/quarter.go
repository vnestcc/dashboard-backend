package models

import (
	"time"

	"gorm.io/gorm"
)

type Quarter struct {
	gorm.Model
	ID        uint `gorm:"primaryKey;autoIncrement"`
	CompanyID uint `gorm:"not null"`
	Date      time.Time
	Quarter   string
	Year      uint

	Company                 Company
	FinancialHealths        []FinancialHealth       `gorm:"foreignKey:QuarterID,references:ID"`
	MarketTractions         []MarketTraction        `gorm:"foreignKey:QuarterID,references:ID"`
	UnitEconomics           []UnitEconomics         `gorm:"foreignKey:QuarterID,references:ID"`
	ProductDevelopments     []ProductDevelopment    `gorm:"foreignKey:QuarterID,references:ID"`
	TeamPerformances        []TeamPerformance       `gorm:"foreignKey:QuarterID,references:ID"`
	FundraisingStatuses     []FundraisingStatus     `gorm:"foreignKey:QuarterID,references:ID"`
	CompetitiveLandscapes   []CompetitiveLandscape  `gorm:"foreignKey:QuarterID,references:ID"`
	OperationalEfficiencies []OperationalEfficiency `gorm:"foreignKey:QuarterID,references:ID"`
	RiskManagements         []RiskManagement        `gorm:"foreignKey:QuarterID,references:ID"`
	AdditionalInfos         []AdditionalInfo        `gorm:"foreignKey:QuarterID,references:ID"`
	SelfAssessments         []SelfAssessment        `gorm:"foreignKey:QuarterID,references:ID"`
	Attachments             []Attachment            `gorm:"foreignKey:QuarterID,references:ID"`
}
