package models

import (
	"fmt"

	"gorm.io/gorm"
)

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

	IsVisible  int `gorm:"default:63"`
	IsEditable int `gorm:"default:63"`
}

func (o *OperationalEfficiency) TableName() string {
	return "operational"
}

// Bit positions: 0 = OperationalChanges, 1 = ImpactMetrics, 2 = OptimizationAreas, 3 = OperationalBottlenecks, 4 = InfrastructureCapacity, 5 = ScalingPlans
func (o *OperationalEfficiency) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"OperationalChanges":     o.OperationalChanges,
			"ImpactMetrics":          o.ImpactMetrics,
			"OptimizationAreas":      o.OptimizationAreas,
			"OperationalBottlenecks": o.OperationalBottlenecks,
			"InfrastructureCapacity": o.InfrastructureCapacity,
			"ScalingPlans":           o.ScalingPlans,
		}
	}

	result := make(map[string]any)
	if o.IsVisible&(1<<0) != 0 {
		result["OperationalChanges"] = o.OperationalChanges
	}
	if o.IsVisible&(1<<1) != 0 {
		result["ImpactMetrics"] = o.ImpactMetrics
	}
	if o.IsVisible&(1<<2) != 0 {
		result["OptimizationAreas"] = o.OptimizationAreas
	}
	if o.IsVisible&(1<<3) != 0 {
		result["OperationalBottlenecks"] = o.OperationalBottlenecks
	}
	if o.IsVisible&(1<<4) != 0 {
		result["InfrastructureCapacity"] = o.InfrastructureCapacity
	}
	if o.IsVisible&(1<<5) != 0 {
		result["ScalingPlans"] = o.ScalingPlans
	}
	return result
}

func (o *OperationalEfficiency) EditableFilter() error {
	var errFields []string

	if o.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "OperationalChanges")
	}
	if o.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "ImpactMetrics")
	}
	if o.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "OptimizationAreas")
	}
	if o.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "OperationalBottlenecks")
	}
	if o.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "InfrastructureCapacity")
	}
	if o.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "ScalingPlans")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
