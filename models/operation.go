package models

import (
	"fmt"

	"gorm.io/gorm"
)

type OperationalEfficiency struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	OperationalChanges     string `json:"operational_changes"`
	ImpactMetrics          string `json:"impact_metrics"`
	OptimizationAreas      string `json:"optimization_areas"`
	OperationalBottlenecks string `json:"operational_bottlenecks"`
	InfrastructureCapacity string `json:"infrastructure_capacity"`
	ScalingPlans           string `json:"scaling_plans"`

	IsVisible  uint8 `gorm:"default:63" json:"-"`
	IsEditable uint8 `gorm:"default:63" json:"-"`
}

func (o *OperationalEfficiency) TableName() string {
	return "operational"
}

func (o *OperationalEfficiency) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"operational_changes",
		"impact_metrics",
		"optimization_areas",
		"operational_bottlenecks",
		"infrastructure_capacity",
		"scaling_plans",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if o.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

// Bit positions: 0 = operational_changes, 1 = impact_metrics, 2 = optimization_areas, 3 = operational_bottlenecks, 4 = infrastructure_capacity, 5 = scaling_plans
func (o *OperationalEfficiency) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":                 o.Version,
			"operational_changes":     o.OperationalChanges,
			"impact_metrics":          o.ImpactMetrics,
			"optimization_areas":      o.OptimizationAreas,
			"operational_bottlenecks": o.OperationalBottlenecks,
			"infrastructure_capacity": o.InfrastructureCapacity,
			"scaling_plans":           o.ScalingPlans,
		}
	}

	result := make(map[string]any)
	result["version"] = o.Version
	if o.IsVisible&(1<<0) != 0 {
		result["operational_changes"] = o.OperationalChanges
	}
	if o.IsVisible&(1<<1) != 0 {
		result["impact_metrics"] = o.ImpactMetrics
	}
	if o.IsVisible&(1<<2) != 0 {
		result["optimization_areas"] = o.OptimizationAreas
	}
	if o.IsVisible&(1<<3) != 0 {
		result["operational_bottlenecks"] = o.OperationalBottlenecks
	}
	if o.IsVisible&(1<<4) != 0 {
		result["infrastructure_capacity"] = o.InfrastructureCapacity
	}
	if o.IsVisible&(1<<5) != 0 {
		result["scaling_plans"] = o.ScalingPlans
	}
	return result
}

func (o *OperationalEfficiency) EditableFilter() error {
	var errFields []string

	if o.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "operational_changes")
	}
	if o.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "impact_metrics")
	}
	if o.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "optimization_areas")
	}
	if o.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "operational_bottlenecks")
	}
	if o.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "infrastructure_capacity")
	}
	if o.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "scaling_plans")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
