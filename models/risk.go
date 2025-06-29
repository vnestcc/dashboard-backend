package models

import (
	"fmt"

	"gorm.io/gorm"
)

type RiskManagement struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	RegulatoryChanges  string `json:"regulatory_changes"`
	ComplianceStatus   string `json:"compliance_status"`
	RegulatoryConcerns string `json:"regulatory_concerns"`
	SecurityAudits     string `json:"security_audits"`
	DataProtection     string `json:"data_protection"`
	SecurityIncidents  string `json:"security_incidents"`
	KeyDependencies    string `json:"key_dependencies"`
	ContingencyPlans   string `json:"contingency_plans"`

	IsVisible  uint8 `gorm:"default:255" json:"-"`
	IsEditable uint8 `gorm:"default:255" json:"-"`
}

func (r *RiskManagement) TableName() string {
	return "risk"
}

func (r *RiskManagement) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"regulatory_changes",
		"compliance_status",
		"regulatory_concerns",
		"security_audits",
		"data_protection",
		"security_incidents",
		"key_dependencies",
		"contingency_plans",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if r.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

// Bit positions: 0 = regulatory_changes, 1 = compliance_status, 2 = regulatory_concerns, 3 = security_audits,
// 4 = data_protection, 5 = security_incidents, 6 = key_dependencies, 7 = contingency_plans
func (r *RiskManagement) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":             r.Version,
			"regulatory_changes":  r.RegulatoryChanges,
			"compliance_status":   r.ComplianceStatus,
			"regulatory_concerns": r.RegulatoryConcerns,
			"security_audits":     r.SecurityAudits,
			"data_protection":     r.DataProtection,
			"security_incidents":  r.SecurityIncidents,
			"key_dependencies":    r.KeyDependencies,
			"contingency_plans":   r.ContingencyPlans,
		}
	}

	result := make(map[string]any)
	result["version"] = r.Version
	if r.IsVisible&(1<<0) != 0 {
		result["regulatory_changes"] = r.RegulatoryChanges
	}
	if r.IsVisible&(1<<1) != 0 {
		result["compliance_status"] = r.ComplianceStatus
	}
	if r.IsVisible&(1<<2) != 0 {
		result["regulatory_concerns"] = r.RegulatoryConcerns
	}
	if r.IsVisible&(1<<3) != 0 {
		result["security_audits"] = r.SecurityAudits
	}
	if r.IsVisible&(1<<4) != 0 {
		result["data_protection"] = r.DataProtection
	}
	if r.IsVisible&(1<<5) != 0 {
		result["security_incidents"] = r.SecurityIncidents
	}
	if r.IsVisible&(1<<6) != 0 {
		result["key_dependencies"] = r.KeyDependencies
	}
	if r.IsVisible&(1<<7) != 0 {
		result["contingency_plans"] = r.ContingencyPlans
	}
	return result
}

func (r *RiskManagement) EditableFilter() error {
	var errFields []string

	if r.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "regulatory_changes")
	}
	if r.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "compliance_status")
	}
	if r.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "regulatory_concerns")
	}
	if r.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "security_audits")
	}
	if r.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "data_protection")
	}
	if r.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "security_incidents")
	}
	if r.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "key_dependencies")
	}
	if r.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "contingency_plans")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
