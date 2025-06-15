package models

import "gorm.io/gorm"

type OperationalEfficiency struct {
	gorm.Model
	CompanyID              uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID              uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version                int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	OperationalChanges     string
	ImpactMetrics          string
	OptimizationAreas      string
	OperationalBottlenecks string
	InfrastructureCapacity string
	ScalingPlans           string

	IsVisible  int
	IsEditable int
}
